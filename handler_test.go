package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("GET",
		"/blog/:year/:month",
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			year := ParamFromContext(ctx, "year")
			month := ParamFromContext(ctx, "month")
			t.Log(year)
			t.Log(month)
			w.Write([]byte("OK"))
		},
	)

	srv := httptest.NewServer(router)

	resp, err := http.Get(srv.URL + "/blog/2020/10")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Log(resp.Status)
}
