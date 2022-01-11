package mux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRouter_HandleFunc(t *testing.T) {
	tt := []struct {
		name       string
		pattern    string
		requestURL string
		wantCalled bool
		wantParams map[string]string
	}{
		{
			name:       "no_match",
			pattern:    "/test",
			requestURL: "/nope",
		},
		{
			name:       "no_dynamic_match",
			pattern:    "/hello/{name}",
			requestURL: "/nope",
		},
		{
			name:       "no_wildcard_match",
			pattern:    "/test/*",
			requestURL: "/nope",
		},
		{
			name:       "no_match_estrict",
			pattern:    "/test/",
			requestURL: "/test",
		},
		{
			name:       "ok",
			pattern:    "/test",
			requestURL: "/test",
			wantCalled: true,
		},
		{
			name:       "ok_one_param",
			pattern:    "/hello/{name}",
			requestURL: "/hello/world",
			wantCalled: true,
			wantParams: map[string]string{
				"name": "world",
			},
		},
		{
			name:       "ok_two_param",
			pattern:    "/file/{name}.{ext}",
			requestURL: "/file/test.txt",
			wantCalled: true,
			wantParams: map[string]string{
				"name": "test",
				"ext":  "txt",
			},
		},
		{
			name:       "ok_wildcard",
			pattern:    "/test/*",
			requestURL: "/test/x",
			wantCalled: true,
		},
		{
			name:       "ok_complex_capture",
			pattern:    "/hello/{name}",
			requestURL: "/hello/a&b c",
			wantCalled: true,
			wantParams: map[string]string{
				"name": "a&b c",
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var gotCalled bool

			router := &Router{}
			router.HandleFunc(tc.pattern, func(w http.ResponseWriter, r *http.Request) {
				gotCalled = true

				var gotParams map[string]string
				for name := range tc.wantParams {
					if gotParams == nil {
						gotParams = map[string]string{}
					}

					val := URLParam(r.Context(), name)
					gotParams[name] = val
				}

				if !reflect.DeepEqual(tc.wantParams, gotParams) {
					t.Errorf("want params %+v; got %+v", tc.wantParams, gotParams)
				}
			})

			srv := httptest.NewServer(router)
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL+tc.requestURL, http.NoBody)
			if err != nil {
				t.Error(err)
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}

			defer resp.Body.Close()

			_, err = io.Copy(io.Discard, resp.Body)
			if err != nil {
				t.Error(err)
				return
			}

			if tc.wantCalled != gotCalled {
				t.Errorf("want to be called %v; got %v", tc.wantCalled, gotCalled)
			}
		})
	}
}

func TestMethodHandler_ServeHTTP(t *testing.T) {
	tt := []struct {
		name          string
		method        string
		requestMethod string
		wantCalled    bool
	}{
		{
			name:          "no_match",
			method:        http.MethodGet,
			requestMethod: http.MethodPost,
		},
		{
			name:          "ok",
			method:        http.MethodGet,
			requestMethod: http.MethodGet,
			wantCalled:    true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var gotCalled bool

			mh := MethodHandler{tc.method: func(w http.ResponseWriter, r *http.Request) {
				gotCalled = true
				w.WriteHeader(http.StatusOK)
			}}
			router := &Router{}
			router.Handle("/", mh)

			srv := httptest.NewServer(router)
			defer srv.Close()

			req, err := http.NewRequest(tc.requestMethod, srv.URL, nil)
			if err != nil {
				t.Error(err)
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}

			defer resp.Body.Close()

			_, err = io.Copy(io.Discard, resp.Body)
			if err != nil {
				t.Error(err)
				return
			}

			if tc.wantCalled != gotCalled {
				t.Errorf("want to be called %v; got %v", tc.wantCalled, gotCalled)
			}
		})
	}
}
