package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
)

func init() {
	tmpl = template.Must(template.ParseGlob("../../view/*.html"))

	// connect to database
	var err error
	db, err = sql.Open("postgres", "postgres://root:password@localhost/codit?sslmode=disable")
	if err != nil {
		log.Println("database open error: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Println("database connect error: ", err)
	}
	log.Println("Connected to the database")
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("../../view/")))

	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/catalogue", catalogueHandler)
	http.HandleFunc("/status", showStatusHandler)
	http.HandleFunc("/userinfo", userInfoHandler)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/problem", problemHandler)

	http.ListenAndServe(":9090", nil)
}
