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
	"strings"
)

var conf struct {
	SrcDir      string
	OutDir      string
	ContentDir  string
	SkelDir     string
	TemplateDir string
}

type Page map[string]interface{}
type Site map[string]interface{}

func Split(s string, d string) []string {
	arr := strings.Split(s, d)
	return arr
}

var helperFuncs = template.FuncMap{
	"Split": Split,
}

func main() {
	conf.SrcDir = "example"
	conf.OutDir = "example/www"
	conf.ContentDir = filepath.Join(conf.SrcDir, "content")
	conf.SkelDir = filepath.Join(conf.SrcDir, "skel")
	conf.TemplateDir = filepath.Join(conf.SrcDir, "templates")
	err := gen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
	}
}

func gen() error {

	// set up the output dir
	err := CopyDir(conf.SkelDir, conf.OutDir)
	if err != nil {
		return err
	}

	tmpls, err := loadTemplates(conf.TemplateDir)
	if err != nil {
		return err
	}

	pages, err := readPages()
	if err != nil {
		return err
	}
	/*
		for _, tmpl := range t.Templates() {
			fmt.Println(tmpl.Tree.Name)
		}
	*/

	// TODO: support a site config file
	site := Site{"Pages": pages}

	// prep the pages for rendering
	for _, page := range pages {
		err := cookPage(page, site)
		if err != nil {
			return err
		}
	}

	// render page content (to "content" key)
	for _, page := range pages {
		err := renderPageContent(page, site)
		if err != nil {
			return err
		}
	}

	// render full pages
	for _, page := range pages {

		err := renderPage(page, site, tmpls)
		if err != nil {
			return err
		}
	}
	return nil
}

func renderPage(page Page, site Site, tmpls *template.Template) error {
	//	fmt.Println(t.DefinedTemplates())

	// TODO: allow per-page templates
	tmplName := "default.html"
	def := tmpls.Lookup(tmplName)
	if def == nil {
		return fmt.Errorf("ERR: missing template '%s'", tmplName)
	}

	// work out output filename
	relPath := page["path"].(string)
	file := page["slug"].(string) + ".html"

	outFilename := filepath.Join(conf.OutDir, relPath, file)

	err := os.MkdirAll(filepath.Join(conf.OutDir, relPath), 0777)
	if err != nil {
		return err
	}
	outFile, err := os.Create(outFilename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	fmt.Printf("----- %s -----\n", page["_srcfile"])
	var data = struct {
		Site Site
		Page Page
	}{
		site,
		page,
	}
	err = def.Execute(outFile, data)
	if err != nil {
		return err
	}
	return nil
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

	return template.New("").Funcs(helperFuncs).ParseFiles(found...)
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
//	page["content"] = template.HTML(rendered)

func readPages() ([]Page, error) {
	pages := []Page{}

	err := filepath.Walk(conf.ContentDir, func(path string, info os.FileInfo, err error) error {
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

	relPath, err := filepath.Rel(conf.ContentDir, filename)
	if err != nil {
		return err
	}

	page["path"] = filepath.Dir(relPath)
	page["slug"] = filepath.Base(relPath)

	page["url"] = relPath
	return nil
}

func renderPageContent(page Page, site Site) error {

	rawContent := page["_rawcontent"].([]byte)
	ext := filepath.Ext(page["_srcfile"].(string))
	if ext == ".md" {
		rendered := blackfriday.MarkdownCommon(rawContent)
		page["content"] = template.HTML(rendered)
	} else if ext == ".html" {
		// TODO: treat as template!
		tmpl := template.New("_foo").Funcs(helperFuncs)
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
		err = tmpl.Execute(&buf, data)
		if err != nil {
			return err
		}

		page["content"] = template.HTML(buf.String())
	} else {
		page["content"] = string(rawContent)
	}
	return nil
}
