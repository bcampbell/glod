package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// FancyDir is a http.Filesystem implementation which
// also tries ".html" extensions if the requested file is not found
type FancyDir string

func (d FancyDir) Open(name string) (http.File, error) {
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) || strings.Contains(name, "\x00") {
		return nil, errors.New("http: invalid character in file path")
	}
	dir := string(d)
	if dir == "" {
		dir = "."
	}

	exts := []string{"", ".html"}

	var err error
	var f *os.File
	for _, ext := range exts {
		f, err = os.Open(filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name+ext))))
		if err == nil {
			return f, nil
		}
	}
	// return the last error
	return nil, err
}

func serveSite(outDir string) error {

	h := http.FileServer(FancyDir(outDir))
	fmt.Fprintf(os.Stderr, "serving site on http://localhost:8080\n")
	return http.ListenAndServe(":8080", h)

}
