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
	"strings"
)

// helper fn for papering over some of the icky casting
func getStr(dict map[string]interface{}, key string) string {
	if val, ok := dict[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// Page variables:
//
// set in front matter:
//   title	  - title of page. If not set, derived from slug.
//   date
//   template - name of template to use to render this page (default: "default.html")
//
// generated (ie treat as read-only)
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
//   pages    A list of all the pages in the site, indexed by "path/slug"
//            Need to use the index fn to get around non alpha-numeric names
//            eg, in a template:
//                <a href="{{(index .Site.pages "docs/getting-started").url}}>Getting Started</a>
//
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
		os.Exit(1)
	}
	if serverFlag {
		outDir := getStr(site, "_outdir")
		go func() {
			for {
				var err error
				err = waitForChanges(site)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
					os.Exit(1)
				}
				site, err = gen(siteDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
				}
			}
		}()

		err = serveSite(outDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
			os.Exit(1)
		}
	}
}

// The main driver function for generating a site
func gen(siteDir string) (Site, error) {

	//
	site, err := loadSiteConfig(siteDir)
	if err != nil {
		return nil, err
	}

	// TODO: skel dir should be optional
	// set up the output dir
	err = CopyDir(getStr(site, "_skeldir"), getStr(site, "_outdir"))
	if err != nil {
		return nil, err
	}

	tmpls, err := loadTemplates(getStr(site, "_templatesdir"))
	if err != nil {
		return nil, err
	}

	pages, err := readContent(site)
	if err != nil {
		return nil, err
	}

	site["pages"] = pages

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

	fmt.Fprintf(os.Stdout, "generated site\n")
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
		if ext := filepath.Ext(path); ext != ".html" {
			// .html files only
			return nil
		}
		found = append(found, path)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("No templates - %s doesn't exist", srcDir)
		}
		return nil, err
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("No templates found in %s", srcDir)
	}
	return template.New("").Funcs(helperFuncs).ParseFiles(found...)
}

// load in all the pages
// TODO: support having static files in here too
func readContent(site Site) (map[string]Page, error) {
	pages := map[string]Page{}

	err := filepath.Walk(getStr(site, "_contentdir"), func(fullpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(fullpath)
		switch strings.ToLower(ext) {
		case ".md", ".markdown", ".html", ".htm":
			page, err := readPage(site, fullpath)
			if err != nil {
				return fmt.Errorf("%s: %s", fullpath, err)
			}
			slug := getStr(page, "slug")
			relPath := getStr(page, "path")
			pages[path.Join(relPath, slug)] = page
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
// for later rendering (filename is absolute path, not content-relative)
func readPage(site Site, filename string) (Page, error) {

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
	relPath, err := filepath.Rel(getStr(site, "_contentdir"), filename)
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

	// ensure date set (even if empty string!)
	if _, ok := page["date"]; !ok {
		page["date"] = ""
	}

	title := getStr(page, "title")
	if title == "" {
		// derive title from slug
		page["title"] = strings.Title(strings.Replace(slug, "-", " ", -1))
	}

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
	switch strings.ToLower(ext) {
	case ".md", ".markdown", ".mdown":
		// TODO: would be nice to pass markdown through a template here... text/template maybe?
		// or maybe pass it through, then use html/template on the result?
		rendered := blackfriday.MarkdownCommon(rawContent)
		page["content"] = template.HTML(rendered)
	case ".html", ".htm":
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
	default:
		page["content"] = string(rawContent)
	}
	return nil
}

// render the whole page
func renderPage(page Page, site Site, tmpls *template.Template) error {
	outDir := getStr(site, "_outdir")

	tmplName := getStr(page, "template")
	if tmplName == "" {
		tmplName = "default.html"
	}

	def := tmpls.Lookup(tmplName)
	if def == nil {
		return fmt.Errorf("%s: missing template '%s'", getStr(page, "_srcfile"), tmplName)
	}

	// work out output filename
	relPath := getStr(page, "path")
	file := getStr(page, "slug") + ".html"
	outFilename := filepath.Join(outDir, relPath, file)

	err := os.MkdirAll(filepath.Join(outDir, relPath), 0777)
	if err != nil {
		return fmt.Errorf("%s: Mkdir failed: %s", getStr(page, "_srcfile"), err)
	}
	outFile, err := os.Create(outFilename)
	if err != nil {
		return fmt.Errorf("%s: Create failed: %s", getStr(page, "_srcfile"), err)
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
		return fmt.Errorf("%s: render failed: %s", getStr(page, "_srcfile"), err)
	}
	//fmt.Fprintf(os.Stderr, "generated %s\n", outFilename)
	return nil
}

func loadSiteConfig(siteDir string) (Site, error) {

	fileName := filepath.Join(siteDir, "config.toml")
	site := Site{}
	_, err := toml.DecodeFile(fileName, &site)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", fileName, err)
	}

	site["_configfile"] = fileName
	site["_outdir"] = filepath.Join(siteDir, "www")
	site["_contentdir"] = filepath.Join(siteDir, "content")
	site["_skeldir"] = filepath.Join(siteDir, "skel")
	site["_templatesdir"] = filepath.Join(siteDir, "templates")

	return site, err
}
