package router_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Code-Hex/go-router-simple"
)

func TestHandler(t *testing.T) {
	r := router.New()

	r.GET("/", http.HandlerFunc(Index))
	r.GET("/hi/:name", http.HandlerFunc(Hi))
	r.GET("/hi/{name}", http.HandlerFunc(Hi))
	r.GET("/download/*.*", http.HandlerFunc(GetFilename))
	r.GET(`/blog/{year:\d{4}}/{month:(?:\d{2})}`, Blog())

	srv := httptest.NewServer(r)

	for _, path := range []string{
		"/",
		"/hi/codehex",
		"/hi/taro",
		"/download/file.xml",
		"/download/filename.json",
		"/blog/2020/10",
	} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("path: %q, code: %d", path, resp.StatusCode)
		}
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Index!\n")
}

func Hi(w http.ResponseWriter, r *http.Request) {
	name := router.ParamFromContext(r.Context(), "name")
	fmt.Fprintf(w, "Welcome! %q\n", name)
}

func GetFilename(w http.ResponseWriter, r *http.Request) {
	wildcards := router.WildcardsFromContext(r.Context())
	fmt.Fprintf(w, "File: %q\n", strings.Join(wildcards, "."))
}

func Blog() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		year := router.ParamFromContext(r.Context(), "year")
		month := router.ParamFromContext(r.Context(), "month")
		fmt.Fprintf(w, "Render: %q\n", year+"/"+month)
	})
}
