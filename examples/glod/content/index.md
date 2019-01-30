# GLOD

Glod is a static site generator, along the lines of [Jekyll](http://jekyllrb.com/)
or [Hugo](https://gohugo.io). But way more spartan than either of those.


## Overview

A Glod site has a directory structure something like this:

    .
    +-- site.toml       overall site config
    +-- skel/           css, js, images and other static files
    +-- templates/      default.html etc...
    +-- content/        .md files passed through templates to produce pages

When glod is run, it performs these steps:

1. load `site.toml` into the `.Site` variable.
2. create a new output directory, `www`.
3. copy everything in `skel` to `www`.
4. for each file in `content`:
    1. read the front matter values into the `.Page` variable.
    2. process the file through a template (`default.html` by default, but can be specified in front matter). The template can use values in `.Page` and `.Site`.
    3. write the resultant HTML output to `www`, preserving the same directory structure as in `content`.

## Invoking


    $ glod [-server] [SITEDIR]

`SITEDIR` is the top-level directory containing `site.toml`. It may be omitted to operate upon the current directory.

`-server` will cause glod to run a local web server to serve the site on. It will appear at `http://localhost:8080`.
In `-server` mode, altering any of the files will cause an automatic rebuild of the site.


## `site.toml`

This holds the overall site configuration.

The values in `site.toml` are available for use in templates through the `.Site` variable.

You can place any site-wide data you like in here, but there are a few names which
have special meanings:

`title`
: The *pretty* name for the site, eg "`Fancy Site`"

`baseurl`
: the full url for the root of the site, eg "`http://glod.scumways.com`"

`pages`
: a list of all the pages in the site, indexed by "path/slug".
This is built when glod is run - you should not set `pages`
in `site.toml`, but it will be available for use in templates
via `.Site`.

an example `site.toml`:

    title = "Glod"
    baseurl = "http://glod.scumways.com/"


## `skel` dir

This holds the base files for the site. It is copied verbatim as the first step in building the site.

Usually you'd use this to hold any static css, js or image files.

## `templates` dir

This holds any templates used to compose content into HTML pages.
`default.html` is the default page template but this can be overriden per-page in the front matter, using the `template` variable.

Templates are written in the [golang template format](https://golang.org/pkg/text/template/).

Templates can be for full HTML pages, or individual page components intended to be included from other templates (eg page headers and footers).

It's up to you to decide how you want to organise them.
You can use organise them in subdirectories, but remember that the template takes it's name from the base filename, without the subdirectory.
If multiple templates have the same name, only one of them will be defined.

For example:

`header.html`:
 
    <!DOCTYPE html>
    <html>
      <head>
        <title>{{.Site.title}} - {{.Page.title}}</title>
        <link href="style.css" rel="stylesheet"/>
        <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
      </head>
    <body>
    <header class="site-header">
    <h1>{{.Site.title}}</h1>
    </header>

`footer.html`:

    </body>
    </html>

`default.html`:

    {{template "header.html" .}}
    <h2>{{.Page.title}}</h2>
    {{.Page.content}}
    {{template "footer.html" .}}



## `content` dir

Each file `content` represents a single page.

Page content files can be markdown (`.md`) or HTML (`.html`) files.

They are processed through a template and rendered out to an `.html` file.

Page-specific values can be defined in a front matter section.
The front matter is denoted by `+++`. For example:

    +++
    title="Fancy blog post"
    date="2011-02-05"
    +++

    Here is content for our fancy blog posting, in markdown.

    ... blah blah ...


The directory structure in `content` is preserved and will be reflected in the rendered website.


HTML files are passed through the template engine, and can access variables.
However the templates from the `templates` directory are /not/ available here.
For example, you could generate an index page like this:

    +++
    title="Index"
    +++
    <h2>Blog posts</h2>
    <ul>
    {{ range .Site.pages }}
    {{ if in .path "blog" }}
    <li><a href="{{.url}}">{{.title}}</a></li>
    {{end}}
    {{end}}
    </ul>

### `.Site` var

Everything in `site.toml` is made available to the template through `.Site`.

### `.Page` var

This holds values specific to the current page being processed.
It is initialised from the front matter section of the content file.

`.Page` is really a shortcut to the entry in `.Site.pages` for
the current page.

You can put any values you like in the front matter.

The names with special meanings:

`title`
: The page title. If not set in the front matter, it is derived from the
slug by replacing hyphens with spaces and title-casing the text, eg "`hello-there`"
becomes "`Hello There`".

`date`
: The timestamp of the page, eg "2016-08-20"
(TODO: what other date/time formats are supported?)

`template`
: name of template to use to render this page (default: `default.html`).

There are a few names which are not to be set in the front matter.
They are derived values, calculated at runtime:

`url`
: The full URL of the page, relative to the site root
  (eg "`/posts/april/everything-is-a-bit-shit`").
  There is a special case: if the slug is `index`, then the url is truncated.
  eg
    * "`index.html`" -> "`/`", 
    * "`posts/index.html`" -> "`/posts/`"

`path`
: The path part of the URL, eg `posts/april`

`slug`
: The slug part of the URL, eg `everything-is-a-bit-shit`

`content`
: holds the rendered (html) content for the page


## Template Helper Functions

Various helper functions available within templates:


### `split [string] [sep]`

Splits `string` wherever `sep` occurs, returning a slice containing the multiple pieces.

### `in [haystack] [needle]`

Returns true if `needle` is found in `haystack`

eg, generate a list of all pages under `posts/`:
```
  <ul>
  {{ range .Site.pages }}{{ if in .path "posts/" }}
    <li>{{ .date }} <a href="{{.url}}">{{.title }}</a></li>
  {{ end }}{{ end }}
  </ul>
```


### `sort [collection] <sortField> <sortOrder>`

sorts collections (maps, arrays or slices) of elements

optional args:

`sortField`
: If the elements contain sub elements, `sortField` lets you pick the one use as a sort key. (eg, you might sorts posts using `"date"`). By default, maps are sorted by key and slices/arrays are sorted by value.

`sortOrder`
: must be "asc" (default) or "desc"

TODO: examples


### `dateFormat [fmt] [date]`

Formats the `date` according to the `fmt` string.

TODO: examples. Document `fmt`.

## URL policy

It's assumed that you have fully control over how your web server maps URLs to pages.

Glod aims to produce nice clean urls like:

    https://example.com/blog/example-post

...without the `.html` suffix or any silly trickery like giving every page it's own directory containing a single `index.html` file.

The assumption is that you set up your web server to handle the pages correctly as html.

TODO: example configs for nginx and apache.

TODO: Should probably support an option for .html extensions or separate directories

## NOTES

### Getting specific pages from `.Site.pages`

Because of non-alphanumeric characters in keys, you'll probably need to use the `index` function in templates. eg:

This won't work because of the '/' and '-':
```
<a href="{{index .Site.pages.docs/getting-started.url}}>Getting Started</a>
```

But this will:
```
<a href="{{(index .Site.pages "docs/getting-started").url}}>Getting Started</a>
```

