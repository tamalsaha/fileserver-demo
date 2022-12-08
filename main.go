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
	dir := flag.String("d", "files", "the directory of static file to host")
	flag.Parse()

	_ = os.MkdirAll(*dir, 0755)

	fileServer := http.FileServer(http.Dir(*dir))

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

	log.Printf("Serving %s on HTTP port: %s\n", *dir, *port)
	log.Fatal(http.ListenAndServe(":"+*port, r))
}

const MaxUploadSize = 100 << 20 // 100 MB

// FileSave fetches the file and saves to disk
func FileSave(dir string, r *http.Request) (string, error) {
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

	size, err := getSize(f)
	if err != nil {
		//// logger.WithError(err).Error("failed to get the size of the uploaded content")
		//w.WriteHeader(http.StatusInternalServerError)
		//writeError(w, err)
		return "", err
	}
	if size > MaxUploadSize {
		// logger.WithField("size", size).Info("file size exceeded")
		// w.WriteHeader(http.StatusRequestEntityTooLarge)
		// writeError(w, errors.New("uploaded file size exceeds the limit"))
		return "", errors.New("uploaded file size exceeds the limit")
	}

	filename := h.Filename
	if filename == "" {
		return "", errors.New("missing file name")
	}

	fullPath := filepath.Join(dir, r.URL.Path, filename)
	_ = os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
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

func getSize(content io.Seeker) (int64, error) {
	size, err := content.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	_, err = content.Seek(0, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return size, nil
}
