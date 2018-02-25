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
	// parse templates
	tmpl = template.Must(template.ParseGlob("../../view/*.html"))

	// connect to database
	db, err = sql.Open("postgres", "postgres://root:password@localhost/codit?sslmode=disable")
	if err != nil {
		log.Println("database open error: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Connected to the database")
	}

	// init variables
	row := db.QueryRow("SELECT rid FROM submissions ORDER BY rid DESC LIMIT 1")
	err = row.Scan(&rid)
	if err != nil {
		log.Println("cannot get last rid - ", err)
	}
}

func main() {
	// write log to file
	logFile, err := os.OpenFile("../../filesystem/logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("func init cannot open log file: ", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

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
	
	log.Fatal(http.ListenAndServe(":9090", nil))
}
