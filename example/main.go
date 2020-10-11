package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Code-Hex/go-router-simple"
	"github.com/julienschmidt/httprouter"
)

func main() {
	r := router.New()
	r.GET(`/blog/{year:\d{4}}/{month:(?:\d{2})}`, Blog())

	// Advanced usage
	// The way of throw most of the routing to httprouter and some to go-router-simple.
	hr := httprouter.New()
	hr.GET("/hi/:name", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		name := ps.ByName("name")
		fmt.Fprintf(w, "Welcome! %q\n", name)
	})
	hr.GET("/bye/:name", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		name := ps.ByName("name")
		fmt.Fprintf(w, "Bye! %q\n", name)
	})

	// If not found on httprouter, it will traverse on go-router-simple.
	hr.NotFound = r

	log.Fatal(http.ListenAndServe(":8080", hr))
}

func Blog() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		year := router.ParamFromContext(r.Context(), "year")
		month := router.ParamFromContext(r.Context(), "month")
		fmt.Fprintf(w, "Render: %q\n", year+"/"+month)
	})
}
