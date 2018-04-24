package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
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
		http.Redirect(w, r, "/catalogue", http.StatusSeeOther)
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
		err := row.Scan(&user.ID, &user.Password, &user.Admin)
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
		log.Println(user.ID + " signed up successfully")
		return
	}
	tmpl.ExecuteTemplate(w, "signup.html", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if alreadyLoggedIn(w, r) {
		http.Redirect(w, r, "/catalogue", http.StatusSeeOther)
		return
	}

	// process form submission
	if r.Method == http.MethodPost {
		un := r.FormValue("username")
		p := r.FormValue("password")

		// is there a username?
		row := db.QueryRow("SELECT * FROM members WHERE id = $1", un)
		user := Member{}
		err := row.Scan(&user.ID, &user.Password, &user.Admin)
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
		log.Println(user.ID + " logged in successfully")
		http.Redirect(w, r, "/catalogue", http.StatusSeeOther)
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
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT pid, title, level FROM problems ORDER BY pid ASC")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Println("func catalogueHandler cannot query problem catalogue in SQL -", err)
		return
	}
	defer rows.Close()

	pbs := make([]ProblemInfo, 0)
	mem := getUser(w, r)
	for rows.Next() {
		pbinfo := ProblemInfo{}

		// get Pid, Title, Level
		var level int
		err := rows.Scan(&pbinfo.Pid, &pbinfo.Title, &level)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			log.Println("func catalogueHandler query problem catalogue in SQL error -", err)
			return
		}
		switch level {
		case 0:
			pbinfo.Level0 = true
		case 1:
			pbinfo.Level1 = true
		case 2:
			pbinfo.Level2 = true
		case 3:
			pbinfo.Level3 = true
		case 4:
			pbinfo.Level4 = true
		}

		// caculate and convert Acceptance to string for display
		var ac, total float64
		row := db.QueryRow("SELECT count(*) FROM submissions WHERE problem = $1 and result = 1", pbinfo.Pid)
		err = row.Scan(&ac)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			log.Println("func catalogueHandler cannot query problems -", err)
			return
		}
		row = db.QueryRow("SELECT count(*) FROM submissions WHERE problem = $1", pbinfo.Pid)
		err = row.Scan(&total)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			log.Println("func catalogueHandler scan problems error -", err)
			return
		}
		pbinfo.Acceptance = strconv.FormatFloat(ac/total*100, 'f', 2, 32)
		if pbinfo.Acceptance != "NaN" {
			pbinfo.Acceptance += "%"
		}

		// get State
		pbinfo.State = false
		if mem.ID != "" {
			var rid int
			row = db.QueryRow("SELECT rid FROM submissions WHERE problem = $1 and result = 1 and username = $2", pbinfo.Pid, mem.ID)
			if row.Scan(&rid) == nil {
				pbinfo.State = true
			}
		}

		pbs = append(pbs, pbinfo)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tmpl.ExecuteTemplate(w, "catalogue.html", pbs)
}

func problemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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
	var description, input, output, sampleInput, sampleOutput string
	err := row.Scan(&pb.Pid, &pb.Title, &description, &input, &output,
		&sampleInput, &sampleOutput, &pb.Level)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		log.Println("func problemHandler problem ", pid, " not found - ", err)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		log.Println("func problemHandler - problem ", pid, " query error - ", err)
		return
	}
	pb.Description = template.HTML(description)
	pb.Input = template.HTML(input)
	pb.Output = template.HTML(output)
	pb.SampleInput = template.HTML(sampleInput)
	pb.SampleOutput = template.HTML(sampleOutput)

	row = db.QueryRow("SELECT count(*) FROM submissions WHERE problem = $1 and result = 1", pb.Pid)
	err = row.Scan(&pb.Accepted)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		log.Println("func problemHandler get total accepted error -", err)
		return
	}
	row = db.QueryRow("SELECT count(*) FROM submissions WHERE problem = $1", pb.Pid)
	err = row.Scan(&pb.Submissions)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		log.Println("func problemHandler get total submissions error -", err)
		return
	}

	tmpl.ExecuteTemplate(w, "problem.html", pb)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	mem := getUser(w, r)
	if mem.ID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// write code to file
	atomic.AddUint64(&rid, 1)
	f, err := os.Create("../../filesystem/submissions/" + strconv.FormatUint(rid, 10) + "." + r.FormValue("compiler"))
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		log.Println("func submitHandler create file error -", err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(r.FormValue("code"))
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		log.Println("func submitHandle write file error -", err)
		return
	}
	f.Sync()

	// get file type
	var ftype int
	switch r.FormValue("compiler") {
	case "c":
		ftype = 0
	case "c++":
		ftype = 1
	}

	// get problem ID
	var pid int
	pid = 1000

	res := operateFile(rid, ftype, 1000)
	_, err = db.Exec("INSERT INTO submissions (rid, username, problem, result, submit_time, language) VALUES ($1, $2, $3, $4, $5, $6)",
		rid, mem.ID, pid, res, time.Now(), ftype)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		log.Println("func submitHandler insert submission info error -", err)
		return
	}

	http.Redirect(w, r, "/status", http.StatusSeeOther)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT * FROM submissions ORDER BY rid DESC LIMIT 20")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Println("func statusHandler cannot query submissions -", err)
		return
	}
	defer rows.Close()

	subs := make([]Submission, 0)
	for rows.Next() {
		sub := Submission{}
		var st time.Time
		var lan int
		var res int

		err := rows.Scan(&sub.RID, &sub.Username, &sub.Problem, &res, &sub.RunTime, &sub.Memory, &st, &lan)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			log.Println("func statusHandler scan submissions erro -", err)
			return
		}

		sub.SubmitTime = st.Format(time.RFC3339)
		switch lan {
		case 0:
			sub.Language0 = true
		case 1:
			sub.Language1 = true
		case 2:
			sub.Language2 = true
		}
		switch res {
		case 0:
			sub.Result0 = true
		case 1:
			sub.Result1 = true
		case 2:
			sub.Result2 = true
		case 3:
			sub.Result3 = true
		case 4:
			sub.Result4 = true
		case 5:
			sub.Result5 = true
		case 6:
			sub.Result6 = true
		case 7:
			sub.Result7 = true
		case 8:
			sub.Result8 = true
		case 9:
			sub.Result9 = true
		}

		subs = append(subs, sub)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tmpl.ExecuteTemplate(w, "status.html", subs)
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		pb := ProblemString{}
		pb.Pid, _ = strconv.Atoi(r.FormValue("pid"))
		pb.Title = r.FormValue("title")
		pb.Level, _ = strconv.Atoi(r.FormValue("level"))
		pb.Description = r.FormValue("description")
		pb.Input = r.FormValue("input")
		pb.Output = r.FormValue("output")
		pb.SampleInput = r.FormValue("sampleinput")
		pb.SampleOutput = r.FormValue("sampleoutput")

		_, err := db.Exec("INSERT INTO problems (pid, title, level, description, input, output, sample_input, sample_output) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			pb.Pid, pb.Title, pb.Level, pb.Description, pb.Input, pb.Output, pb.SampleInput, pb.SampleOutput)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			log.Println("func editHandler db update error - ", err)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	// get page or problem info
	if r.Method == http.MethodGet {
		pid := r.FormValue("pid")
		if pid == "" {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		// get problem info from db
		row := db.QueryRow("SELECT * FROM problems WHERE pid = $1", pid)

		pb := ProblemString{}
		err := row.Scan(&pb.Pid, &pb.Title, &pb.Description, &pb.Input, &pb.Output,
			&pb.SampleInput, &pb.SampleOutput, &pb.Level)
		switch {
		case err == sql.ErrNoRows:
			http.NotFound(w, r)
			log.Println("func editHandler problem ", pid, " not found - ", err)
			return
		case err != nil:
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			log.Println("func editHandler problem ", pid, " query error - ", err)
			return
		}

		// encode to JSON
		response, err := json.Marshal(pb)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			log.Println("func editHandler JSON marshal err - ", err)
			return
		}
		fmt.Fprintf(w, string(response))
	}

	// post form
	if r.Method == http.MethodPost {
		pb := ProblemString{}
		pb.Pid, _ = strconv.Atoi(r.FormValue("pid"))
		pb.Title = r.FormValue("title")
		pb.Level, _ = strconv.Atoi(r.FormValue("level"))
		pb.Description = r.FormValue("description")
		pb.Input = r.FormValue("input")
		pb.Output = r.FormValue("output")
		pb.SampleInput = r.FormValue("sampleinput")
		pb.SampleOutput = r.FormValue("sampleoutput")

		_, err := db.Exec("UPDATE problems SET title = $1, level = $2, description = $3, input = $4, output = $5, sample_input = $6, sample_output = $7 WHERE pid = $8",
			pb.Title, pb.Level, pb.Description, pb.Input, pb.Output, pb.SampleInput, pb.SampleOutput, pb.Pid)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			log.Println("func editHandler db update error - ", err)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func dashHandler(w http.ResponseWriter, r *http.Request) {
	if isAdmin(w, r) == false {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	tmpl.ExecuteTemplate(w, "dashboard.html", nil)
}

// TODO: display history code in page
func codeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rid := r.FormValue("rid")
	if rid == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
}
