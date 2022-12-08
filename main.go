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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	hw "go.wandrs.dev/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func main() {
	port := flag.String("p", "8100", "port to serve on")
	dir := flag.String("d", "files", "the directory of static file to host")
	flag.Parse()

	prefix := "files" // where files are stored
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	_ = os.MkdirAll(*dir, 0o755)
	fileServer := http.FileServer(http.Dir(*dir))

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	pattern := path.Join(prefix, "*")

	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.Options(pattern, http.StripPrefix(prefix, fileServer).ServeHTTP)
	r.Get(pattern, http.StripPrefix(prefix, fileServer).ServeHTTP)
	r.Post(pattern, func(w http.ResponseWriter, r *http.Request) {
		err := FileSave(prefix, *dir, r)

		status := hw.ErrorToAPIStatus(err)
		code := int(status.Code)
		// when writing an error, check to see if the status indicates a retry after period
		if status.Details != nil && status.Details.RetryAfterSeconds > 0 {
			delay := strconv.Itoa(int(status.Details.RetryAfterSeconds))
			w.Header().Set("Retry-After", delay)
		}
		if code == http.StatusNoContent {
			w.WriteHeader(code)
			return
		}
		data, _ := json.MarshalIndent(status, "", "  ")
		_, _ = w.Write(data)
	})

	// http.Handle("/", fileServer)

	log.Printf("Serving %s on HTTP port: %s\n", *dir, *port)
	log.Fatal(http.ListenAndServe(":"+*port, r))
}

const MaxUploadSize = 100 << 20 // 100 MB

// FileSave fetches the file and saves to disk
func FileSave(prefix, dir string, r *http.Request) error {
	// left shift 100 << 20 which results in 32*2^20 = 33554432
	// x << y, results in x*2^y
	// 1 MB max memory
	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		return err
	}
	// Retrieve the file from form data
	f, h, err := r.FormFile("file")
	if err != nil {
		return err
	}
	defer f.Close()

	size, err := getSize(f)
	if err != nil {
		//// logger.WithError(err).Error("failed to get the size of the uploaded content")
		//w.WriteHeader(http.StatusInternalServerError)
		//writeError(w, err)
		return err
	}
	if size > MaxUploadSize {
		// logger.WithField("size", size).Info("file size exceeded")
		// w.WriteHeader(http.StatusRequestEntityTooLarge)
		// writeError(w, errors.New("uploaded file size exceeds the limit"))
		return apierrors.NewRequestEntityTooLargeError(fmt.Sprintf("received %s, limit %s", humanize.Bytes(uint64(size)), humanize.Bytes(MaxUploadSize)))
	}

	filename := h.Filename
	if filename == "" {
		return errors.New("missing file name")
	}

	fullPath := filepath.Join(dir, strings.TrimPrefix(r.URL.Path, prefix), filename)
	_ = os.MkdirAll(filepath.Dir(fullPath), 0o755)
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	// Copy the file to the destination path
	_, err = io.Copy(file, f)
	if err != nil {
		return err
	}
	return nil
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
