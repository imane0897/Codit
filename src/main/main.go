package main

import (
	"database/sql"
	"html/template"
	"log"
	// "net/http"
	"os"

	"github.com/buaazp/fasthttprouter"
	_ "github.com/lib/pq"
	"github.com/valyala/fasthttp"
)

func init() {
	// parse templates
	tmpl = template.Must(template.ParseGlob("../../view/*.html"))

	// connect to database
	db, err = sql.Open("postgres", "postgres://aym:password@localhost/codit?sslmode=disable")
	if err != nil {
		log.Println("database open error: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Connected to the database")
	}

	// init global variable rid
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

	router := fasthttprouter.New()
	router.ServeFiles("/view/*filepath", "/Users/aym/Documents/Codit/view")
	router.GET("/", homeHandler)
	router.GET("/signup", signupHandler)
	router.POST("/signup", signupPostHandler)
	router.GET("/login", loginHandler)
	router.POST("/login", loginPostHandler)
	router.GET("/logout", logoutHandler)
	router.GET("/catalogue", catalogueHandler)
	router.GET("/status", statusHandler)
	// router.GET("/submit", submitHandler)
	router.GET("/problem", problemHandler)
	router.GET("/dashboard", dashHandler)
	router.GET("/editproblem", editHandler)
	// router.POST("/newproblem", newHandler)
	router.GET("/userinfo", getUserInfo)
	router.GET("/pidcount", getPidCount)

	log.Fatal(fasthttp.ListenAndServe(":9090", router.Handler))
}
