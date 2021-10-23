package mux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestMux_HandleFunc(t *testing.T) {
	tt := []struct {
		name          string
		method        string
		requestMethod string
		pattern       string
		requestURL    string
		wantCalled    bool
		wantParams    map[string]string
	}{
		{
			method:        http.MethodGet,
			requestMethod: http.MethodGet,
			pattern:       "/foo",
			requestURL:    "/bar",
		},
		{
			method:        http.MethodGet,
			requestMethod: http.MethodPost,
			pattern:       "/foo",
			requestURL:    "/foo",
		},
		{
			method:        "*",
			requestMethod: http.MethodPost,
			pattern:       "/foo",
			requestURL:    "/foo",
			wantCalled:    true,
		},
		{
			method:        "*",
			requestMethod: http.MethodPost,
			pattern:       "/foo",
			requestURL:    "/bar",
		},
		{
			method:        http.MethodPost,
			requestMethod: http.MethodPost,
			pattern:       "/hello/{name}",
			requestURL:    "/hello/world",
			wantCalled:    true,
			wantParams: map[string]string{
				"name": "world",
			},
		},
		{
			method:        http.MethodPost,
			requestMethod: http.MethodGet,
			pattern:       "/hello/{name}",
			requestURL:    "/hello/world",
		},
		{
			method:        http.MethodPatch,
			requestMethod: http.MethodPatch,
			pattern:       "/foo/{foo}/baz/{baz}",
			requestURL:    "/foo/bar/baz/qux",
			wantCalled:    true,
			wantParams: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
		},
		{
			method:        "*",
			requestMethod: http.MethodPatch,
			pattern:       "/foo/{foo}/baz/{baz}",
			requestURL:    "/foo/bar/baz/qux",
			wantCalled:    true,
			wantParams: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
		},
		{
			method:        http.MethodDelete,
			requestMethod: http.MethodDelete,
			pattern:       "/foo/{foo}/baz/{baz}",
			requestURL:    "/foo/bar/baz/qux",
			wantCalled:    true,
			wantParams: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
		},
		{
			method:        http.MethodOptions,
			requestMethod: http.MethodOptions,
			pattern:       "/foo/{foo}",
			requestURL:    "/foo/bar/baz",
		},
		{
			method:        http.MethodHead,
			requestMethod: http.MethodHead,
			pattern:       "/foo/*",
			requestURL:    "/baz/qux",
		},
		{
			method:        http.MethodPut,
			requestMethod: http.MethodPut,
			pattern:       "/hello/{name}",
			requestURL:    "/hello/a&b c",
			wantCalled:    true,
			wantParams: map[string]string{
				"name": "a&b c",
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var gotCalled bool

			mux := &Mux{}
			mux.HandleFunc(tc.method, tc.pattern, func(w http.ResponseWriter, r *http.Request) {
				gotCalled = true

				if want, got := tc.method, r.Method; want != got && want != "*" {
					t.Errorf("want method %q; got %q", want, got)
					return
				}

				if tc.wantParams != nil {
					gotParams := make(map[string]string)
					for name := range tc.wantParams {
						val := URLParam(r.Context(), name)
						gotParams[name] = val
					}

					if !reflect.DeepEqual(tc.wantParams, gotParams) {
						t.Errorf("want params %+v; got %+v", tc.wantParams, gotParams)
					}
				}
			})

			srv := httptest.NewServer(mux)
			defer srv.Close()

			req, err := http.NewRequest(tc.requestMethod, srv.URL+tc.requestURL, nil)
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
			}

			if tc.wantCalled != gotCalled {
				t.Errorf("want to be called %v; got %v", tc.wantCalled, gotCalled)
			}
		})
	}
}
