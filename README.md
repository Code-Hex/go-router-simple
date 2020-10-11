go-router-simple
=====
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Code-Hex/go-router-simple)](https://pkg.go.dev/github.com/Code-Hex/go-router-simple)

go-router-simple is a simple HTTP request router for [Go](https://golang.org/). This package is ported from [Router::Simple](https://metacpan.org/pod/Router::Simple) which is a great module in Perl.

**Motivation:** Most request routing third-party modules in Go are implemented using the trie (radix) tree algorithm. Hence, they are fast but it hard to provide flexible URL parameters. (e.g. URL paths containing `/`)

⚠️ **Supports Go1.12 and above.**

## Synopsis

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Code-Hex/go-router-simple"
)

func main() {
	r := router.New()

	r.GET("/", http.HandlerFunc(Index))
	r.GET("/hi/:name", http.HandlerFunc(Hi))
	r.GET("/hi/{name}", http.HandlerFunc(Hi))
	r.GET("/download/*.*", http.HandlerFunc(GetFilename))
	r.GET(`/blog/{year:\d{4}}/{month:(?:\d{2})}`, Blog())

	log.Fatal(http.ListenAndServe(":8080", r))
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
```

## Advanced usage

go-router-simple is used regular expressions for traversal but it's a little slow. So you would better to use other third-party routing packages for traversing most request paths whenever possible.

Here's example. Basically, use httprouter, but use go-router-simple for complex URL parameters.

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Code-Hex/go-router-simple"
	"github.com/julienschmidt/httprouter"
)

func main() {
    // Setup go-router-simple.
	r := router.New()
	r.GET(`/blog/{year:(?:199\d|20\d{2})}/{month:(?:0?[1-9]|1[0-2])}`, Blog())

	// Setup httprouter.
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
```