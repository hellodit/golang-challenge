package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte // it's mean a byte slice
}

// Global variable to call once at program initiations
var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html", "tmpl/index.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// This is a method named save that takes as its receiver p, a pointer to Page .
// It takes no parameters, and returns a value of type error
//  If all goes well, Page.save() will return nil

func (p *Page) save() error {
	os.Mkdir("data", 0777)
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.

}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, _ := ioutil.ReadFile(filename)
	return &Page{
		Title: title,
		Body:  body,
	}, nil

}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}

	renderTemplate(w, "edit", p)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title: "Welcome to the jungle",
		Body:  []byte("Hallo met datang ya"),
	}
	renderTemplate(w, "index", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {

	body := r.FormValue("body")

	p := &Page{
		Title: title,
		Body:  []byte(body),
	}

	err := p.save()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.HandleFunc("/", RootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
