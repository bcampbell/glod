package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type Page map[string]interface{}
type Site map[string]interface{}

func main() {

	t, err := loadTemplates("example/templates")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
		os.Exit(1)
	}

	pages, err := readPages("example/content")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
		os.Exit(1)
	}
	/*
		for _, tmpl := range t.Templates() {
			fmt.Println(tmpl.Tree.Name)
		}
	*/

	site := Site{"Pages": pages}
	for _, page := range pages {
		err := cookPage(page, site)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
			os.Exit(1)
		}
	}
	for _, page := range pages {
		err := renderPage(page, site)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
			os.Exit(1)
		}
	}

	for _, page := range pages {
		//	fmt.Println(t.DefinedTemplates())
		var data = struct {
			Site Site
			Page Page
		}{
			site,
			page,
		}

		def := t.Lookup("default.html")
		if def == nil {
			fmt.Fprintf(os.Stderr, "ERR: missing template\n")
			os.Exit(1)
		}

		fmt.Printf("----- %s -----\n", page["_srcfile"])
		err = def.Execute(os.Stdout, data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
			os.Exit(1)
		}
	}
}

func loadTemplates(srcDir string) (*template.Template, error) {

	found := []string{}

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		found = append(found, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return template.ParseFiles(found...)
}

var tomlFrontMatterPat = regexp.MustCompile(`(?ms)\A[+]{3}\s*$\s*(.*?)^[+]{3}\s*$\s*(.*)\z`)

func readPage(filename string) (Page, error) {

	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	m := tomlFrontMatterPat.FindSubmatch(raw)

	//	var in struct{ Title, Date string }
	page := Page{}
	content := []byte{}
	if m == nil {
		content = raw
	} else {
		_, err = toml.Decode(string(m[1]), &page)
		if err != nil {
			return nil, err
		}
		content = m[2]
	}

	// stash content for later rendering
	page["_rawcontent"] = content
	page["_srcfile"] = filename

	return page, nil
}

// render it
//	rendered := blackfriday.MarkdownCommon(content)
//	page["Content"] = template.HTML(rendered)

func readPages(srcDir string) ([]Page, error) {
	pages := []Page{}

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		switch ext {
		case ".md", ".html":
			page, err := readPage(path)
			if err != nil {
				return err
			}
			pages = append(pages, page)
		default:
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pages, nil
}

func cookPage(page Page, site Site) error {
	filename := page["_srcfile"].(string)

	page["URL"] = filename

	return nil
}

// TODO: make templates and funcs available here
func renderPage(page Page, site Site) error {

	rawContent := page["_rawcontent"].([]byte)
	ext := filepath.Ext(page["_srcfile"].(string))
	if ext == ".md" {
		rendered := blackfriday.MarkdownCommon(rawContent)
		page["Content"] = template.HTML(rendered)
	} else if ext == ".html" {
		// TODO: treat as template!
		tmpl := template.New("_foo")
		_, err := tmpl.Parse(string(rawContent))
		if err != nil {
			return err
		}

		data := struct {
			Page Page
			Site Site
		}{
			page, site,
		}

		var buf bytes.Buffer
		tmpl.Execute(&buf, data)

		page["Content"] = template.HTML(buf.String())
	} else {
		page["Content"] = string(rawContent)
	}
	return nil
}
