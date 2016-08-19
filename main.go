package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

// TODO: kill this. use site vars instead
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
//   template - name of template to use to render this page (default: "default.html")
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

func main() {
	flag.Usage = func() {

		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "%s [OPTIONS] [SITEDIR]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generate a static website from templates.\n")
		flag.PrintDefaults()
	}

	var serverFlag bool
	flag.BoolVar(&serverFlag, "server", false, "run webserver after generating site")
	flag.Parse()

	siteDir := "."
	if flag.NArg() > 0 {
		siteDir = flag.Arg(0)
	}

	site, err := gen(siteDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
	}

	if serverFlag {
		err = serveSite(site)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
		}
	}
}

// The main driver function for generating a site
func gen(siteDir string) (Site, error) {

	conf.SrcDir = siteDir
	conf.OutDir = filepath.Join(conf.SrcDir, "www")
	conf.ContentDir = filepath.Join(conf.SrcDir, "content")
	conf.SkelDir = filepath.Join(conf.SrcDir, "skel")
	conf.TemplateDir = filepath.Join(conf.SrcDir, "templates")

	//
	site, err := loadSiteConfig(conf.SrcDir)
	if err != nil {
		return nil, err
	}

	// TODO: skel dir should be optional
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

// load and parse all the templates in the templates dir
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

// load in all the pages
// TODO: support having static files in here too
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

var tomlFrontMatterPat = regexp.MustCompile(`(?ms)\A[+]{3}\s*$\s*(.*?)^[+]{3}\s*$\s*(.*)\z`)

// read a page from the content dir, parsing the front matter and stashing the raw content
// for later rendering
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

	// add various computed variables
	relPath, err := filepath.Rel(conf.ContentDir, filename)
	if err != nil {
		return nil, err
	}

	relDir := filepath.Dir(relPath)
	page["path"] = relDir

	slug := stripExt(filepath.Base(relPath))
	page["slug"] = slug

	var u string
	if slug == "index" {
		u = path.Join("/", relDir, "/")
	} else {
		u = path.Join("/", relDir, slug)
	}
	page["url"] = u

	return page, nil
}

func stripExt(path string) string {
	for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			return path[:i]
		}
	}
	return path
}

// add various computed variables to a page
func cookPage(page Page, site Site) error {
	//filename := page["_srcfile"].(string)

	return nil
}

// render the raw content stashed in a page and store it in the "content" field
func renderPageContent(page Page, site Site) error {

	rawContent := page["_rawcontent"].([]byte)
	ext := filepath.Ext(page["_srcfile"].(string))
	if ext == ".md" {
		// TODO: would be nice to pass markdown through a template here... text/template maybe?
		// or maybe pass it through, then use html/template on the result?
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
	tmplName, ok = page["template"].(string)
	if !ok || tmplName == "" {
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
	fmt.Fprintf(os.Stderr, "generated %s\n", outFilename)
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
