package mux

import "testing"

func Test_patternToRegExp(t *testing.T) {
	tt := []struct {
		name  string
		given string
		want  string
	}{
		{
			name:  "simple",
			given: "test",
			want:  "^test$",
		},
		{
			name:  "simple_segments",
			given: "foo/bar",
			want:  "^foo/bar$",
		},
		{
			name:  "named_capture",
			given: "{capture}",
			want:  `^(?P<capture>[^\/]+)$`,
		},
		{
			name:  "named_capture_segment",
			given: "/foo/{bar}/baz/{qux}",
			want:  `^/foo/(?P<bar>[^\/]+)/baz/(?P<qux>[^\/]+)$`,
		},
		{
			name:  "wildcard",
			given: "*",
			want:  "^(.*)$",
		},
		{
			name:  "wildcard_segment",
			given: "/test/*",
			want:  "^/test/(.*)$",
		},
		{
			name:  "named_capture_and_wildcard",
			given: "/foo/{bar}/*",
			want:  `^/foo/(?P<bar>[^\/]+)/(.*)$`,
		},
		{
			name:  "wildcard_and_named_capture",
			given: "/*/foo/{bar}",
			want:  `^/(.*)/foo/(?P<bar>[^\/]+)$`,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := patternToRegExp(tc.given).String()
			if tc.want != got {
				t.Errorf("patternToRegExp(%q) = %q, want %q", tc.given, got, tc.want)
			}
		})
	}
}
