package main

import (
	"database/sql"
	"html/template"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB
var tmpl *template.Template
var rid uint64
var err error

// Member has the information to interact with DATABASE codit TABLE members
// id in 3-10 char, pwd in 6-10 char.
// -----PostgreSQL-----
// id   |   varchar(30)
// pwd  |   bytea
type Member struct {
	ID       string
	Password []byte
	Admin    bool
}

// Submission has the info for table display in status.html
// --------------PostgreSQL-----------------
// rid         | serial8
// username    | varchar(10)
// problem     | int
// result      | int
// run_time    | int
// memory      | int
// submit_time | timestamp without time zone
// language    | int
// -----------------------------------------
// Result0		Pending
// Result1		Accept
// Result2		Wrong Answer
// Result3		Compile Error
// Result4		Runtime Error
// Result5		Time Limit Exceeded
// Result6		Memory Limit Exceeded
// Result7		Output Limit Exceeded
// Result8		Presentation Error
// Result9		System Error
// Language0	C
// Language1	C++
// Language2	Java
type Submission struct {
	RID        int
	Username   string
	Problem    int
	Result0    bool
	Result1    bool
	Result2    bool
	Result3    bool
	Result4    bool
	Result5    bool
	Result6    bool
	Result7    bool
	Result8    bool
	Result9    bool
	RunTime    int
	Memory     int
	SubmitTime string
	Language0  bool
	Language1  bool
	Language2  bool
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
// level         |      int
type Problem struct {
	Pid          int
	Title        string
	Description  template.HTML
	Input        template.HTML
	Output       template.HTML
	SampleInput  template.HTML
	SampleOutput template.HTML
	Level        int
}

// ProblemInfo has the info for table in catalogue.html
type ProblemInfo struct {
	Pid        int
	Title      string
	Acceptance string
	Level0     bool
	Level1     bool
	Level2     bool
	Level3     bool
	Level4     bool
	State      bool
}

// ProblemString is used to convey problem info in JSON format
type ProblemString struct {
	Pid          int
	Title        string
	Description  string
	Input        string
	Output       string
	SampleInput  string
	SampleOutput string
	Level        int
}
