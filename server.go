package main

import (
	"errors"
	//	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//
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

func runSite(site Site) error {

	h := http.FileServer(FancyDir(conf.OutDir))
	return http.ListenAndServe(":8080", h)

}
