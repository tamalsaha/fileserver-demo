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
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	r.Post("/*", func(w http.ResponseWriter, r *http.Request) {
		path, err := FileSave(r)
		if path == "" {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(path))
	})

	// http.Handle("/", fileServer)

	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, r))
}

// FileSave fetches the file and saves to disk
func FileSave(r *http.Request) (string, error) {
	// left shift 100 << 20 which results in 32*2^20 = 33554432
	// x << y, results in x*2^y
	// 1 MB max memory
	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		return "", err
	}
	// Retrieve the file from form data
	f, h, err := r.FormFile("file")
	if err != nil {
		return "", err
	}
	defer f.Close()

	filename := h.Filename
	if filename == "" {
		return "", errors.New("missing file name")
	}

	path := filepath.Join(".", "files")
	_ = os.MkdirAll(path, os.ModePerm)
	fullPath := path + "/" + filename
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// Copy the file to the destination path
	_, err = io.Copy(file, f)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}
