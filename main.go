/*
Serve is a very simple static file server in go
Usage:

	-p="8100": port to serve on
	-d=".":    the directory of static files to host

Navigating to http://localhost:8100 will display the index.html or directory
listing file.
*/
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	port := flag.String("p", "8100", "port to serve on")
	directory := flag.String("d", ".", "the directory of static file to host")
	flag.Parse()

	fileServer := http.FileServer(http.Dir(*directory))

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Options("/*", fileServer.ServeHTTP)
	r.Get("/*", fileServer.ServeHTTP)

	// http.Handle("/", fileServer)

	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, r))
}
