package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Code-Hex/go-router-simple"
	"github.com/julienschmidt/httprouter"
)

const (
	wantName = "codehex"
	path     = "/hi/codehex"
)

func request(b *testing.B, h http.Handler) {
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if http.StatusOK != w.Code {
		b.Fatalf("want %q, but got %q", http.StatusOK, w.Code)
	}
}

func BenchmarkHTTPRouter(b *testing.B) {
	hr := httprouter.New()
	hr.GET("/hi/:name", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if got := p.ByName("name"); wantName != got {
			b.Fatalf("want %q, but got %q", wantName, got)
		}
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request(b, hr)
	}
}

func BenchmarkRouterSimple(b *testing.B) {
	r := router.New()
	r.GET("/hi/:name", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		got := router.ParamFromContext(req.Context(), "name")
		if wantName != got {
			b.Fatalf("want %q, but got %q", wantName, got)
		}
	}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request(b, r)
	}
}
