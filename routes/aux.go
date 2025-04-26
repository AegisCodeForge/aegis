package routes

import (
	"fmt"
	"log"
	"net/http"
	"html/template"
)

func LogIfError(err error) {
	if err != nil {
		log.Print(err.Error())
	}
}

// go don't have ufcs so i'll have to suffer.
func WithLog(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(fmt.Sprintf(" %s %s", r.Method, r.URL.Path))
		f(w, r)
	}
}
func WithLogHandler(f http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(fmt.Sprintf(" %s %s", r.Method, r.URL.Path))
		f.ServeHTTP(w, r)
	}
}

func FoundAt(w http.ResponseWriter, p string) {
	w.Header().Add("Content-Length", "0")
	w.Header().Add("Location", p)
	w.WriteHeader(302)
}

func LoadTemplate(t *template.Template, name string) *template.Template {
	res := t.Lookup(name)
	if res == nil { log.Fatal(fmt.Sprintf("Failed to find template \"%s\"", name)) }
	return res
}

func LogTemplateError(e error) {
	if e != nil { log.Print(e) }
}

