package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"os"
	"path"
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

// Page variables:
//
// set in front matter:
//   title
//   date
//
// generated:
//   path     - eg "posts/april"
//   slug     - eg "everything-is-a-bit-shit"
//   url      - eg "/posts/april/everything-is-a-bit-shit"
//   content  - holds the rendered content for page
//
type Page map[string]interface{}

// Site variables:
//
//   title    (eg "Fancy Site")
//   baseurl  (eg "http://fancysite.example.com/") TODO
//   publishdir  (default: "www") TODO
//   uglyurls  (default: false) TODO

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
	site, err := gen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
	}

	err = runSite(site)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
	}

}

func gen() (Site, error) {

	site, err := loadSiteConfig(conf.SrcDir)
	if err != nil {
		return nil, err
	}

	// set up the output dir
	err = CopyDir(conf.SkelDir, conf.OutDir)
	if err != nil {
		return nil, err
	}

	tmpls, err := loadTemplates(conf.TemplateDir)
	if err != nil {
		return nil, err
	}

	pages, err := readContent()
	if err != nil {
		return nil, err
	}
	/*
		for _, tmpl := range t.Templates() {
			fmt.Println(tmpl.Tree.Name)
		}
	*/

	site["Pages"] = pages

	// prep the pages for rendering
	for _, page := range pages {
		err := cookPage(page, site)
		if err != nil {
			return nil, err
		}
	}

	// render page content (to "content" key)
	for _, page := range pages {
		err := renderPageContent(page, site)
		if err != nil {
			return nil, err
		}
	}

	// render full pages
	for _, page := range pages {

		err := renderPage(page, site, tmpls)
		if err != nil {
			return nil, err
		}
	}
	return site, err
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

func readContent() ([]Page, error) {
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

func stripExt(path string) string {
	for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			return path[:i]
		}
	}
	return path
}

func cookPage(page Page, site Site) error {
	filename := page["_srcfile"].(string)

	relPath, err := filepath.Rel(conf.ContentDir, filename)
	if err != nil {
		return err
	}

	relDir := filepath.Dir(relPath)
	page["path"] = relDir

	slug := stripExt(filepath.Base(relPath))
	page["slug"] = slug

	page["url"] = path.Join("/", relDir, slug) + ".html"
	return nil
}

// set the "content" field on a page
func renderPageContent(page Page, site Site) error {

	rawContent := page["_rawcontent"].([]byte)
	ext := filepath.Ext(page["_srcfile"].(string))
	if ext == ".md" {
		rendered := blackfriday.MarkdownCommon(rawContent)
		page["content"] = template.HTML(rendered)
	} else if ext == ".html" {
		tmpl := template.New("").Funcs(helperFuncs)
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

// render the whole page
func renderPage(page Page, site Site, tmpls *template.Template) error {
	//	fmt.Println(t.DefinedTemplates())

	var tmplName string
	var ok bool
	if tmplName, ok = page["template"].(string); !ok {
		tmplName = "default.html"
	}

	def := tmpls.Lookup(tmplName)
	if def == nil {
		return fmt.Errorf("missing template '%s'", tmplName)
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

func loadSiteConfig(siteDir string) (Site, error) {

	fileName := filepath.Join(siteDir, "config.toml")
	site := Site{}
	_, err := toml.DecodeFile(fileName, &site)
	if err != nil {
		return nil, err
	}

	return site, err
}
