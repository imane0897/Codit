package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func init() {
	// write log to file
	LogFile, err := os.OpenFile("../../filesystem/logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("func init cannot open log file: ", err)
	}
	defer LogFile.Close()
	log.SetOutput(LogFile)

	// parse templates
	tmpl = template.Must(template.ParseGlob("../../view/*.html"))

	// connect to database
	db, err = sql.Open("postgres", "postgres://root:password@localhost/codit?sslmode=disable")
	if err != nil {
		log.Println("database open error: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Println("database connect error: ", err)
	} else {
		log.Println("Connected to the database")
	}

	// init variables
	row := db.QueryRow("SELECT rid FROM submissions LIMIT 1")
	err = row.Scan(&rid)
	if err != nil {
		log.Println("cannot get last rid: ", err)
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("../../view/")))

	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/catalogue", catalogueHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/userinfo", userInfoHandler)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/problem", problemHandler)
	http.HandleFunc("/code", codeHandler)

	http.ListenAndServe(":9090", nil)
}
