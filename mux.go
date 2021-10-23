package mux

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

var keyURLParams = struct{ name string }{name: "url-params"}

type Mux struct {
	NotFoundHandler         func() http.Handler
	MethodNotAllowedHandler func(allow []string) http.Handler

	once          sync.Once
	staticRoutes  map[string][]route
	dynamicRoutes map[string][]route
}

// New returns a new Mux.
func New() *Mux {
	return &Mux{}
}

type route struct {
	method   string
	pathname string
	re       *regexp.Regexp
	handler  http.Handler
}

func (mux *Mux) init() {
	mux.once.Do(func() {
		if mux.NotFoundHandler == nil {
			mux.NotFoundHandler = http.NotFoundHandler
		}
		if mux.MethodNotAllowedHandler == nil {
			mux.MethodNotAllowedHandler = func(allow []string) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Allow", strings.Join(allow, ", "))
					w.WriteHeader(http.StatusMethodNotAllowed)
				})
			}
		}

		mux.staticRoutes = map[string][]route{}
		mux.dynamicRoutes = map[string][]route{}
	})
}

// Handle regosters a handler for the given method and pattern.
func (mux *Mux) Handle(method, pattern string, handler http.Handler) {
	mux.init()

	if !isPattern(pattern) {
		mux.staticRoutes[pattern] = append(mux.staticRoutes[pattern], route{
			method:   method,
			pathname: pattern,
			handler:  handler,
		})
		return
	}

	re := patternToRegExp(pattern)
	mux.dynamicRoutes[pattern] = append(mux.dynamicRoutes[pattern], route{
		method:  method,
		re:      re,
		handler: handler,
	})
}

// Handle regosters a handler function for the given method and pattern.
func (mux *Mux) HandleFunc(method, pattern string, handler http.HandlerFunc) {
	mux.Handle(method, pattern, handler)
}

// ServeHTTP dispatches the request to the handler whose method and pattern matches,
// otherwise it responds with "not found" or "method not allowed" accordingly.
func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := cleanPath(r.URL.Path)
	routes, ok := mux.staticRoutes[path]
	if ok {
		var allow []string
		for _, route := range routes {
			if route.method != r.Method && route.method != "*" {
				allow = append(allow, route.method)
				continue
			}

			route.handler.ServeHTTP(w, r)
			return
		}

		if allow != nil {
			mux.MethodNotAllowedHandler(allow).ServeHTTP(w, r)
			return
		}
	}

	for _, routes := range mux.dynamicRoutes {
		var allow []string
		for _, route := range routes {
			matches := route.re.FindAllStringSubmatch(path, -1)
			if matches == nil {
				break // all other routes in this map entry will have the same regexp.
			}

			if route.method != r.Method && route.method != "*" {
				allow = append(allow, route.method)
				continue
			}

			groups := route.re.SubexpNames()
			params := map[string]string{}
			for _, match := range matches {
				for i, value := range match {
					if i >= len(groups) {
						continue
					}

					name := groups[i]
					if name == "" {
						continue
					}

					params[name] = value
				}
			}

			if len(params) != 0 {
				route.handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), keyURLParams, params)))
				return
			}

			route.handler.ServeHTTP(w, r)
			return
		}

		if allow != nil {
			mux.MethodNotAllowedHandler(allow).ServeHTTP(w, r)
			return
		}
	}

	mux.NotFoundHandler().ServeHTTP(w, r)
}

// URLParam extracts an URL parameter previously defined in the URL pattern.
func URLParam(ctx context.Context, name string) string {
	params, ok := ctx.Value(keyURLParams).(map[string]string)
	if !ok {
		return ""
	}

	return params[name]
}
