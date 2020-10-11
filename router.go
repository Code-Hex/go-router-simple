// Because I don't want to use regexp.Copy https://golang.org/pkg/regexp/#Regexp.Copy
//
// +build go1.12

package router

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

const (
	pattern1 = `\{((?:\{[0-9,]+\}|[^{}]+)+)\}` // /blog/{year:\d{4}}
	pattern2 = `:([A-Za-z0-9_]+)`              // /blog/:year
	pattern3 = `(\*)`                          // /blog/*/*
	pattern4 = `([^{:*]+)`                     // normal string

	wildcardKey = "__splat__"
)

// errdidnotmatch represents did not match to any regex
var errdidnotmatch = errors.New("did not match")

var (
	re = regexp.MustCompile(
		strings.Join([]string{
			pattern1,
			pattern2,
			pattern3,
			pattern4,
		}, "|"),
	)

	paramsPool = sync.Pool{
		New: func() interface{} {
			return newParams()
		},
	}
)

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	// ErrorLog logs error in ServeHTTP. If not specified, it defaults
	// to log.Printf is used.
	ErrorLog Logger

	// NotFound is configurable http.Handler which is called when no matching
	// route is found. If it is not set, http.NotFound is used.
	NotFound http.Handler

	pathRegexp *regexp.Regexp
	methodMap  map[string][]regexpCapture
}

// New creates new Router struct.
func New() *Router {
	return &Router{
		pathRegexp: re,
		methodMap:  make(map[string][]regexpCapture, 9), // num methods
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
			if err != errdidnotmatch {
				r.logf("ServeHTTP error: %q", err)
			}
			continue
		}
		defer putParams(params)
		ctx := contextWithParams(req.Context(), params)
		rc.handler.ServeHTTP(w, req.WithContext(ctx))
		return
	}

	// Handle 404
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}

type regexpCapture struct {
	rg       *regexp.Regexp
	captures []string
	handler  http.Handler
}

func (rc regexpCapture) MatchPath(path string) (*params, error) {
	regex := rc.rg
	matches := regex.FindStringSubmatch(path)
	if len(matches) == 0 {
		return nil, errdidnotmatch
	}
	matches = matches[1:]
	if len(rc.captures) > 0 && len(matches) != len(rc.captures) {
		// Should not contain parenthesis in regexp pattern
		//
		// Good: "/{date:(?:\d+)}"
		// Bad:  "/{date:(\d+)}"
		return nil, fmt.Errorf("parameter mismatch with regexp: %q", regex.String())
	}
	// NOTE(codehex): I guess better use sync.Pool
	params := getParams()
	for i, capture := range rc.captures {
		if capture == wildcardKey {
			params.wildcards = append(params.wildcards, matches[i])
		} else {
			params.capture[capture] = matches[i]
		}
	}
	return params, nil
}

func createPathMatcher(path string) (regex string, captures []string) {
	var b strings.Builder
	b.WriteString("^")
	submatches := re.FindAllStringSubmatch(path, -1)
	for _, submatch := range submatches {
		for idx, match := range submatch[1:] {
			if match != "" {
				pattern, capture := replacePattern(idx, match)
				if capture != "" {
					captures = append(captures, capture)
				}
				b.WriteString(pattern)
			}
		}
	}
	b.WriteString("$")
	return b.String(), captures
}

func replacePattern(index int, s string) (string, string) {
	switch index {
	case 0:
		sep := strings.SplitN(s, ":", 2)
		if len(sep) != 2 {
			return "([^/]+)", s
		}
		name, pattern := sep[0], sep[1]
		return "(" + pattern + ")", name
	case 1:
		return "([^/]+)", s
	case 2:
		return "(.+)", wildcardKey
	}
	return regexp.QuoteMeta(s), ""
}

type params struct {
	wildcards []string
	capture   map[string]string
}

func newParams() *params {
	return &params{
		wildcards: make([]string, 0),
		capture:   make(map[string]string, 0),
	}
}

func (p *params) reset() {
	wildcards := p.wildcards
	p.wildcards = wildcards[0:0]
	for key := range p.capture {
		delete(p.capture, key)
	}
}

func getParams() *params {
	ps := paramsPool.Get().(*params)
	ps.reset()
	return ps
}

func putParams(ps *params) {
	if ps != nil {
		paramsPool.Put(ps)
	}
}
