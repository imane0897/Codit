package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

var db *sql.DB

// Member has the information to interact with DATABASE codit TABLE members
// id in 3-10 char, pwd in 6-10 char.
type Member struct {
	ID       string
	Password []byte
}

// Submission has the information to interact with DATABASE codit TABLE submissions
// NOTE that data type of submitTime in db is timestamp or say time.Time in Golang
type Submission struct {
	RID        int
	Username   string
	Problem    int
	Result     int
	RunTime    int
	Memory     int
	SubmitTime time.Time
	Language   int
}

// ShowSubmission has the same struct of Submission but changed datatype of SubmitTime
// and Language for template show use
type ShowSubmission struct {
	RID        int
	Username   string
	Problem    int
	Result     string
	RunTime    int
	Memory     int
	SubmitTime string
	Language   string
}

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://root:password@localhost/codit?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	log.Println("Connected to database codit")
}

func _insert() {
	_, err := db.Exec("INSERT INTO submissions (username, problem, language, result, run_time, memory, submit_time) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		"root", 1000, 1, 2, 0, 313, time.Now())
	if err != nil {
		panic(err)
	}
	var expTime time.Time
	expTime = time.Now()
	log.Println(expTime)
}

func _select() {
	rows, err := db.Query("SELECT * FROM submissions;")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	subs := make([]Submission, 0)
	for rows.Next() {
		sub := Submission{}
		err := rows.Scan(&sub.RID, &sub.Username, &sub.Problem, &sub.Result, &sub.RunTime, &sub.Memory, &sub.SubmitTime, &sub.Language) // order matters
		if err != nil {
			panic(err)
		}
		subs = append(subs, sub)
	}
	fmt.Println(subs)
}
