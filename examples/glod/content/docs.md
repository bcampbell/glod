# GLOD

Glod is a static site generator, along the lines of [Jekyll](http://jekyllrb.com/)
or [Hugo](https://gohugo.io). But way more spartan than either of those.


## Overview

A Glod site has a directory structure something like this:

    .
    +-- config.toml     overall site config
    +-- skel/           css, js, images and other static files
    +-- templates/      default.html etc...
    +-- content/        .md files passed through templates to produce pages

When glod is run, it performs these steps:

1. create a new output directory, `www`.
2. copy everything in `skel` to `www`.
3. each file in the `content` dir is processed through a template, and the resulting output is added to the site. Glod will preserve the stucture you have in `content`.



### `config.toml`

This holds the overall site configuration, plus any extra values you decide to set.

These are all available in templates under the `.Site` value.


eg:

    title = "Glod"
    baseurl = "http://glod.scumways.com/"


See the `.Site` variable for details.





### `skel` dir

This holds the base files for the site. It is copied verbatim as the first step in building the site.

Usually you'd use this to hold any static css, js or image files.

### `templates` dir

This holds any templates used to compose content into HTML pages.
`default.html` is the default page template but this can be overriden per-page in the front matter.

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



### `content` dir

Each file in the content dir is processed and rendered out to an .html file.

The directory structure is preserved, and will be reflected in the rendered website.

Page content files can be markdown (`.md`) or HTML (`.html`) files.
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


Each content file can have a front matter section which defines various values.

    +++
    title="Fancy blog post"
    date="2011-02-05"
    +++

    Here is a fancy blog posting.

    It has content and stuff.




url
: eg "/posts/april/everything-is-a-bit-shit"
  as a special case, if the slug is `index`, then the url is truncated.
  eg "index.html" -> "/"
     "posts/index.html" -> "/posts/"

content
: holds the rendered (html) content for page


## Variables


### `.Site` vars

`.Site` holds values which are relevant to the site as a whole.
It is initialised via `config.toml`, so you can add any
site-global values you'd like.

Keys with special meaning are:


`title`
: The /pretty/ name for the site, eg "`Fancy Site`"

`baseurl`
: the full url for the root of the site, eg "`http://glod.scumways.com`"

`pages`
: a list of all the pages in the site, indexed by "path/slug".
This is calculated when glod is run - you should not set `pages`
in `config.toml`.
Because of non-alphanumeric characters in keys, you'll probably need to use the `index` function in templates. eg:

```
<a href="{{(index .Site.pages "docs/getting-started").url}}>Getting Started</a>
```



### `.Page` vars

This holds values specific to the current page being processed.
Any values set in the front matter of content files shows up here.
`.Page` is really a shortcut to the entry in `.Site.pages` for
the current page.

Values with special meanings:

`title`
: The page title. If not set in the front matter, it is derived from the
slug, by replacing hyphens with spaces and title-casing the text, eg "`hello-there`"
becomes "`Hello There`".

`date`
: The timestamp of the page, eg "2016-08-20"

`template`
: name of template to use to render this page (default: `default.html`). It's
unlikely you'd refernce this from a template, but it's noted here for
completeness.

`url`
: The full URL of the page, relative to the site root.
  eg "/posts/april/everything-is-a-bit-shit"
  as a special case, if the slug is `index`, then the url is truncated.
  eg "index.html" -> "/"
     "posts/index.html" -> "/posts/"


`path`
: The path part of the URL, eg `posts/april`

`slug`
: The slug part of the URL, eg `everything-is-a-bit-shit`

`content`
: holds the rendered (html) content for page



Note: `url`, `path`, `slug` and `content` are not to be set in the front matter.
They are derived values, and should be considered read-only.


## Functions

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



TODO document:

* commandline syntax and -server mode
* URL policy (ie server support required)
* examples (index page)



