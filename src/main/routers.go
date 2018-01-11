package main

import (
	"os"
	// "fmt"
	"net/http"
	"path/filepath"
	"html/template"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	fs := http.FileServer(http.Dir("/html"))
	r.Handle("/", fs)

	cwd, _ := os.Getwd()
	// fmt.Println(filepath.Join(cwd + "/../../html/index.html"))

	tmpl0 := template.Must(template.ParseFiles(filepath.Join(cwd + "/../../html/index.html")))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl0.Execute(w, nil)
	})

	tmpl1 := template.Must(template.ParseFiles(filepath.Join(cwd + "/../../html/login.html")))
	r.HandleFunc("/login.html", func(w http.ResponseWriter, r *http.Request) {
		tmpl1.Execute(w, nil)
	})

	tmpl2 := template.Must(template.ParseFiles(filepath.Join(cwd + "/../../html/signup.html")))
	r.HandleFunc("/signup.html", func(w http.ResponseWriter, r *http.Request) {
		tmpl2.Execute(w, nil)
	})

	tmpl3 := template.Must(template.ParseFiles(filepath.Join(cwd + "/../../html/catalogue.html")))
	r.HandleFunc("/catalogue.html", func(w http.ResponseWriter, r *http.Request) {
		tmpl3.Execute(w, nil)
	})

	tmpl4 := template.Must(template.ParseFiles(filepath.Join(cwd + "/../../html/1000.html")))
	r.HandleFunc("/1000.html", func(w http.ResponseWriter, r *http.Request) {
		tmpl4.Execute(w, nil)
	})


	http.ListenAndServe(":9090", r)
}
