## Quick start

    .
    +-- config.toml     overall site config
    +-- skel/           css, js, images and other static files
    +-- templates/      default.html etc...
    +-- content/        .md files passed through templates to produce pages


Glod builds a site by using the `skel` dir as a base.

On top of this, each file in the `content` dir is processed through a template, and the resulting output is added to the site.

You arrange the directory structure how you like. Glod will preserve the stucture you have in `content`.


### `config.toml`

This holds the overall site configuration, plus any extra values you decide to set.

These are all available in templates under the `.Site` value.


eg:

    title = "Glod"
    baseurl = "http://glod.scumways.com/"

title
: the name of the site

baseurl
: the full url for the root of the site


generated vars - these are available in `.Site`, but shouldn't be set in `config.toml`:

pages
: all the pages in the site, indexed by "path/slug".
Because of non-alphanumeric characters in keys, you'll probably need to use the `index` function in templates. eg:
    <a href="{{(index .Site.pages "docs/getting-started").url}}>Getting Started</a>



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

TODO: example

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

The directory structure here is preserved, so what you have here will be reflected in the rendered website.

Page content files can be markdown (`.md`) or HTML (`.html`) files.
HTML files are treated as templates, and can access variables. However the templates from the `templates` directory are /not/ available here.


Each content file can have a front matter section which defines various variables:


##### set in front matter:

title
: title of page. If not set, derived from slug.

date
: eg "2016-08-20"

template
: name of template to use to render this page (default: "`default.html`")

##### generated (ie treat as read-only)

path
: eg "posts/april"

slug
: eg "everything-is-a-bit-shit"

url
: eg "/posts/april/everything-is-a-bit-shit"
  as a special case, if the slug is `index`, then the url is truncated.
  eg "index.html" -> "/"
     "posts/index.html" -> "/posts/"

content
: holds the rendered (html) content for page



TODO document:

* commandline syntax and -server mode
* URL policy (ie server support required)
* template functions
* examples (index page)



