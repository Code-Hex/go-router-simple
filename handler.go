package router

import (
	"context"
	"log"
	"net/http"
	"regexp"
)

// Router :
type Router struct {
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

// Handle handles handler
func (r *Router) Handle(method, path string, handler http.Handler) error {
	regex, captures := createPathMatcher(path)
	rg, err := regexp.Compile(regex)
	if err != nil {
		return err
	}
	r.methodMap[method] = append(r.methodMap[method], regexpCapture{
		rg:       rg,
		captures: captures,
		handler:  handler,
	})
	return nil
}

// HandleFunc handles http.HandlerFunc
func (r *Router) HandleFunc(method, path string, handler func(http.ResponseWriter, *http.Request)) error {
	return r.Handle(method, path, http.HandlerFunc(handler))
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	rcs := r.methodMap[req.Method]
	for _, rc := range rcs {
		params, err := rc.MatchPath(path)
		if err != nil {
			if err != Errdidnotmatch {
				log.Printf("error: %q", err)
			}
			continue
		}
		ctx := contextWithParams(req.Context(), params)
		rc.handler.ServeHTTP(w, req.WithContext(ctx))
		return
	}
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
