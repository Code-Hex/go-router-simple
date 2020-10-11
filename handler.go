// +build go1.12

package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

// Router :
type Router struct {
	ErrorLog Logger

	pathRegexp *regexp.Regexp
	methodMap  map[string][]regexpCapture
}

// NewRouter creates new Router struct.
func NewRouter() *Router {
	return &Router{
		pathRegexp: re,
		methodMap:  make(map[string][]regexpCapture, 6),
	}
}

func (r *Router) logf(format string, args ...interface{}) {
	if r.ErrorLog != nil {
		r.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	rcs := r.methodMap[req.Method]
	for _, rc := range rcs {
		params, err := rc.MatchPath(path)
		if err != nil {
			if err != Errdidnotmatch {
				r.logf("ServeHTTP error: %q", err)
			}
			continue
		}
		ctx := contextWithParams(req.Context(), params)
		rc.handler.ServeHTTP(w, req.WithContext(ctx))
		return
	}
}

// Handle handles handler
func (r *Router) Handle(method, path string, handler http.Handler) {
	regex, captures := createPathMatcher(path)
	rg, err := regexp.Compile(regex)
	if err != nil {
		msg := fmt.Sprintf(
			`regexp: Compile(%s): error=%q, path=%q`,
			regexp.QuoteMeta(regex), err, path,
		)
		panic(msg)
	}
	r.methodMap[method] = append(r.methodMap[method], regexpCapture{
		rg:       rg,
		captures: captures,
		handler:  handler,
	})
}

// HandleFunc handles http.HandlerFunc
func (r *Router) HandleFunc(method, path string, handler func(http.ResponseWriter, *http.Request)) {
	r.Handle(method, path, http.HandlerFunc(handler))
}

// GET is a shorthand for router.Handle(http.MethodGet, path, handle)
func (r *Router) GET(path string, handler http.Handler) {
	r.Handle(http.MethodGet, path, handler)
}

// HEAD is a shorthand for router.Handle(http.MethodHead, path, handle)
func (r *Router) HEAD(path string, handler http.Handler) {
	r.Handle(http.MethodHead, path, handler)
}

// POST is a shorthand for router.Handle(http.MethodPost, path, handle)
func (r *Router) POST(path string, handler http.Handler) {
	r.Handle(http.MethodPost, path, handler)
}

// PUT is a shorthand for router.Handle(http.MethodPut, path, handle)
func (r *Router) PUT(path string, handler http.Handler) {
	r.Handle(http.MethodPut, path, handler)
}

// PATCH is a shorthand for router.Handle(http.MethodPatch, path, handle)
func (r *Router) PATCH(path string, handler http.Handler) {
	r.Handle(http.MethodPatch, path, handler)
}

// DELETE is a shorthand for router.Handle(http.MethodDelete, path, handle)
func (r *Router) DELETE(path string, handler http.Handler) {
	r.Handle(http.MethodDelete, path, handler)
}

// CONNECT is a shorthand for router.Handle(http.MethodConnect, path, handle)
func (r *Router) CONNECT(path string, handler http.Handler) {
	r.Handle(http.MethodConnect, path, handler)
}

// OPTIONS is a shorthand for router.Handle(http.MethodOptions, path, handle)
func (r *Router) OPTIONS(path string, handler http.Handler) {
	r.Handle(http.MethodOptions, path, handler)
}

// TRACE is a shorthand for router.Handle(http.MethodTrace, path, handle)
func (r *Router) TRACE(path string, handler http.Handler) {
	r.Handle(http.MethodTrace, path, handler)
}

type paramsKey struct{}

func contextWithParams(ctx context.Context, p *params) context.Context {
	return context.WithValue(ctx, paramsKey{}, p)
}

// ParamFromContext gets URL parameters from reqeust context.Context.
func ParamFromContext(ctx context.Context, key string) string {
	p, ok := ctx.Value(paramsKey{}).(*params)
	if !ok {
		return ""
	}
	return p.capture[key]
}

// WildcardsFromContext gets URL wildcard parameters from reqeust context.Context.
func WildcardsFromContext(ctx context.Context) []string {
	p, ok := ctx.Value(paramsKey{}).(*params)
	if !ok {
		return nil
	}
	return p.wildcards
}
