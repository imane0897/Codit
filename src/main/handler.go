package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	mem := getUser(w, r)
	if mem.ID == "" {
		tmpl.ExecuteTemplate(w, "index.html", nil)
	} else {
		tmpl.ExecuteTemplate(w, "catalogue.html", nil)
	}
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if alreadyLoggedIn(w, r) {
		tmpl.ExecuteTemplate(w, "catalogue.html", nil)
		return
	}

	// process form submission
	if r.Method == http.MethodPost {
		// get form values
		un := r.FormValue("username")
		p := r.FormValue("password")

		// username taken?
		row := db.QueryRow("SELECT * FROM members WHERE id = $1", un)
		user := Member{}
		err := row.Scan(&user.ID, &user.Password)
		if err != sql.ErrNoRows {
			http.Error(w, "Username already taken", http.StatusForbidden)
			return
		}

		// create session
		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		_, err = db.Exec("INSERT INTO sessions (uuid, username, last_activity) VALUES ($1, $2, $3)",
			c.Value, un, time.Now())
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		// store user in dbUsers
		bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		_, err = db.Exec("INSERT INTO members (id, pwd) VALUES ($1, $2)", un, bs)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		// redirect
		user.ID = un
		http.Redirect(w, r, "/catalogue", http.StatusSeeOther)
		log.Println(user.ID + "signed up successfully")
		return
	}
	tmpl.ExecuteTemplate(w, "signup.html", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if alreadyLoggedIn(w, r) {
		tmpl.ExecuteTemplate(w, "catalogue.html", nil)
		return
	}

	// process form submission
	if r.Method == http.MethodPost {
		un := r.FormValue("username")
		p := r.FormValue("password")

		// is there a username?
		row := db.QueryRow("SELECT * FROM members WHERE id = $1", un)
		user := Member{}
		err := row.Scan(&user.ID, &user.Password)
		if err == sql.ErrNoRows {
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}

		// does the entered password match the stored password?
		err = bcrypt.CompareHashAndPassword(user.Password, []byte(p))
		if err != nil {
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}

		// create session
		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		_, err = db.Exec("INSERT INTO sessions (uuid, username, last_activity) VALUES ($1, $2, $3)",
			c.Value, un, time.Now())
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		// redirect
		http.Redirect(w, r, "/catalogue", http.StatusSeeOther)
		log.Println(user.ID + " logged in successfully")
		return
	}
	tmpl.ExecuteTemplate(w, "login.html", nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if !alreadyLoggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	c, _ := r.Cookie("session")

	// delete the session
	_, err := db.Exec("DELETE FROM sessions WHERE uuid = $1", c.Value)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// remove the cookie
	c = &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, c)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
	log.Println("logged out" + c.Value)
}

func catalogueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT * FROM problems ORDER BY pid ASC")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Println("cannot query problem catalogue in SQL")
		return
	}
	defer rows.Close()

	pbs := make([]ProblemInfo, 0)
	for rows.Next() {
		pb := Problem{}
		pbinfo := ProblemInfo{}

		// check Pid, Title, Level
		err := rows.Scan(&pbinfo.Pid, &pbinfo.Title, &pb.Description, &pb.Input, &pb.Output,
			&pb.SampleInput, &pb.SampleOutput, &pbinfo.Level)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			log.Println("query problem catalogue in SQL error")
			return
		}

		// caculate and convert Acceptance to string for display
		var ac, total float64
		row := db.QueryRow("SELECT count(*) FROM submissions WHERE problem = $1 and result = 1", pbinfo.Pid)
		err = row.Scan(&ac)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			log.Println("func catalogueHandler error: cannot query problems")
		}
		row = db.QueryRow("SELECT count(*) FROM submissions WHERE problem = $1", pbinfo.Pid)
		err = row.Scan(&total)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			log.Println("func catalogueHandler error: scan problems error")
		}
		pbinfo.Acceptance = strconv.FormatFloat(ac/total*100, 'f', 2, 32)
		if pbinfo.Acceptance != "NaN" {
			pbinfo.Acceptance += "%"
		}

		// check State
		var rid int
		mem := getUser(w, r)
		row = db.QueryRow("SELECT rid FROM submissions WHERE problem = $1 and result = 1 and username = $2", pbinfo.Pid, mem.ID)
		if row.Scan(&rid) == nil {
			pbinfo.State = true
		}

		pbs = append(pbs, pbinfo)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tmpl.ExecuteTemplate(w, "catalogue.html", pbs)
}

func userInfoHandler(w http.ResponseWriter, r *http.Request) {
	mem := getUser(w, r)
	fmt.Fprintln(w, mem.ID)
}

func problemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	pid := r.FormValue("pid")
	if pid == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM problems WHERE pid = $1", pid)

	pb := Problem{}
	var SampleInput, SampleOutput string
	err := row.Scan(&pb.Pid, &pb.Title, &pb.Description, &pb.Input, &pb.Output,
		&SampleInput, &SampleOutput, &pb.Level)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		log.Println("func problemHandler error: problem ", pid, " not found")
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		log.Println("func problemHandler error: problem ", pid, " query error")
	}
	pb.SampleInput = template.HTML(SampleInput)
	pb.SampleOutput = template.HTML(SampleOutput)

	tmpl.ExecuteTemplate(w, "problem.html", pb)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	mem := getUser(w, r)
	// LOG
	log.Println("turned into submit func", mem.ID)
	log.Println(r.FormValue("compiler"))

	f, err := os.Create("judge/1000/c/main.c")
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer f.Close()
	_, err = f.WriteString(r.FormValue("code"))
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	f.Sync()

	http.Redirect(w, r, "/status", http.StatusSeeOther)
}
