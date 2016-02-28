package main

import (
	"fmt"
	"regexp"
	"io/ioutil"
	"html/template"
	"net/http"
	"path/filepath"
	"github.com/microcosm-cc/bluemonday"
	"github.com/shurcooL/github_flavored_markdown"
)

type Page struct {
	Title string
	Body []byte
}

var templatesDir = "templates"
var includesDir = "includes"
var templateFormat = ".html"
var dataDir = "data"
var dataFormat = ".md"
var templates map[string]*template.Template
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var frontPage = "FrontPage"

func (p *Page) save() error {
	filename := p.Title + dataFormat
	return ioutil.WriteFile(filepath.Join(dataDir, filename), p.Body, 0600)
}

func (p *Page) HtmlBody() template.HTML {
	unsafe := github_flavored_markdown.Markdown(p.Body)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	return template.HTML(html)
}

func loadPage(title string) (*Page, error) {
	filename := title + dataFormat
	body, err := ioutil.ReadFile(filepath.Join(dataDir, filename))
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) error{
	template, ok := templates[tmpl]
	if  !ok {
		return fmt.Errorf("The template %s does not exist.", tmpl)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := template.ExecuteTemplate(w, "base" , p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/" + title, http.StatusFound)
		return
	}
	renderTemplate(w, "view.html", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}

	renderTemplate(w, "edit.html", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/" + frontPage, http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)

		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r , m[2])
	}
}

func main() {

	if templates == nil {
        templates = make(map[string]*template.Template)
    }

	layouts, err := filepath.Glob(templatesDir + "/*" + templateFormat)
	if err != nil {
		fmt.Errorf(err.Error())
	}

	includes, err := filepath.Glob(includesDir + "/*" + templateFormat)
	if err != nil {
		fmt.Errorf(err.Error())
	}

	for _, layout := range layouts {
		files := append(includes, layout)
		templates[filepath.Base(layout)] = template.Must(template.ParseFiles(files...))
	}


	fs := http.FileServer(http.Dir("static"))
	http.HandleFunc("/", rootHandler)
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	http.ListenAndServe(":8000", nil)
}
