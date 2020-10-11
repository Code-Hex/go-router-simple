// Because I don't want to use regexp.Copy https://golang.org/pkg/regexp/#Regexp.Copy
//
// +build go1.12

package router

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

const (
	pattern1 = `\{((?:\{[0-9,]+\}|[^{}]+)+)\}` // /blog/{year:\d{4}}
	pattern2 = `:([A-Za-z0-9_]+)`              // /blog/:year
	pattern3 = `(\*)`                          // /blog/*/*
	pattern4 = `([^{:*]+)`                     // normal string

	wildcardKey = "__splat__"
)

// Errdidnotmatch represents did not match to any regex
var Errdidnotmatch = errors.New("did not match")

var re = regexp.MustCompile(
	strings.Join([]string{
		pattern1,
		pattern2,
		pattern3,
		pattern4,
	}, "|"),
)

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

type regexpCapture struct {
	rg       *regexp.Regexp
	captures []string
	handler  http.Handler
}

func (rc regexpCapture) MatchPath(path string) (*params, error) {
	regex := rc.rg
	matches := regex.FindStringSubmatch(path)
	if len(matches) == 0 {
		return nil, Errdidnotmatch
	}
	matches = matches[1:]
	// Should not contain parenthesis in regexp pattern
	//
	// Good: "/{date:(?:\d+)}"
	// Bad:  "/{date:(\d+)}"
	if len(rc.captures) > 0 && len(matches) != len(rc.captures) {
		return nil, fmt.Errorf("parameter mismatch with regexp: %q", regex.String())
	}
	// NOTE(codehex): I guess better use sync.Pool
	params := newParams()
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
