package mux

import (
	"context"
	"embed"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

var keyURLParams = struct{ name string }{name: "url-params"}

type Router struct {
	NotFoundHandler http.Handler

	once          sync.Once
	staticRoutes  map[string]http.Handler
	dynamicRoutes []dynamicRoute
}

// NewRouter returns a new Router.
func NewRouter() *Router {
	return &Router{}
}

type dynamicRoute struct {
	re      *regexp.Regexp
	handler http.Handler
}

func (r *Router) init() {
	mux.once.Do(func() {
		if r.NotFoundHandler == nil {
			r.NotFoundHandler = http.NotFoundHandler()
		}

		r.staticRoutes = map[string]http.Handler{}
		r.dynamicRoutes = []dynamicRoute{}
	})
}

// Handle registers a handler for the given pattern.
func (r *Router) Handle(pattern string, handler http.Handler) {
	r.init()

	if !isPattern(pattern) {
		r.staticRoutes[pattern] = handler
		return
	}

	re := patternToRegExp(pattern)
	r.dynamicRoutes = append(r.dynamicRoutes, dynamicRoute{
		re:      re,
		handler: handler,
	})
}

// Handle registers a handler function for the given pattern.
func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	r.Handle(pattern, handler)
}

// ServeHTTP dispatches the request to the handler whose pattern matches,
// otherwise it responds with "not found".
func (r *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := cleanPath(r.URL.Path)

	if handler, ok := r.staticRoutes[path]; ok {
		handler.ServeHTTP(w, r)
		return
	}

	for _, route := range mux.dynamicRoutes {
		matches := route.re.FindAllStringSubmatch(path, -1)
		if matches == nil {
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

	r.NotFoundHandler.ServeHTTP(w, r)
}

// URLParam extracts an URL parameter previously defined in the URL pattern.
func URLParam(ctx context.Context, name string) string {
	params, ok := ctx.Value(keyURLParams).(map[string]string)
	if !ok {
		return ""
	}

	return params[name]
}

// MethodHandler maps each handler to the corresponding method.
// Responds with "method not allowed" if none match.
type MethodHandler map[string]http.HandlerFunc

// ServeHTTP dispatches the request to the handler whose method matches,
// otherwise it responds with "method not allowed".
func (mh MethodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	allow := make([]string, 0, len(mh))
	for method, handler := range mh {
		if method == r.Method {
			handler(w, r)
			return
		}

		allow = append(allow, method)
	}

	w.Header().Set("Allow", strings.Join(allow, ", "))
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}
