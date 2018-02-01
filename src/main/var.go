package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"html/template"
	"time"
)

var db *sql.DB
var tmpl *template.Template

// Member has the information to interact with DATABASE codit TABLE members
// id in 3-10 char, pwd in 6-10 char.
// -----PostgreSQL-----
// id   |   varchar(30)
// pwd  |   bytea
type Member struct {
	ID       string
	Password []byte
}

// Submission has the information to interact with DATABASE codit TABLE submissions
// NOTE that data type of submitTime in db is timestamp or say time.Time in Golang
// --------------PostgreSQL-----------------
// rid         | serial8
// username    | varchar(10)
// problem     | int
// result      | int
// run_time    | int
// memory      | int
// submit_time | timestamp without time zone
// language    | int
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

// Session has uuid of cookie and username to record its owner
// ----------------PostgreSQL----------------
// uuid          | char(36)
// username      | varchar(10)
// last_activity | timestamp without time zone
type Session struct {
	uuid         string
	username     string
	lastActivity time.Time
}

// Problem has the informaion of each problems
// ---------PostgreSQL---------
// pid           |      int
// title         |      text
// description   |      text
// input         |      text
// output        |      text
// sample_input  |      text
// sample_output |      text
type Problem struct {
	Pid          int
	Title        string
	Description  string
	Input        string
	Output       string
	SampleInput  template.HTML
	SampleOutput template.HTML
	Level        int
}

// ProblemInfo has the info for table in catalogue.html 
type ProblemInfo struct {
	Pid        int
	Title      string
	Acceptance string
	Level      int
	State      bool
}
