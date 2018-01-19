package main

import (
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("../../html")))
	
	http.HandleFunc("/sign-up", signup)
	http.HandleFunc("/log-in", login)
	http.HandleFunc("/log-out", logout)
	http.HandleFunc("/status.html", showStatus)
	
	http.HandleFunc("/userinfo", userinfo)
	http.HandleFunc("/submit", submit)

	http.ListenAndServe(":9090", nil)
}

func signup(w http.ResponseWriter, r *http.Request) {
	if alreadyLoggedIn(w, r) {
		http.Redirect(w, r, "/catalogue.html", http.StatusSeeOther)
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
		_, err = db.Exec("INSERT INTO sessions (uuid, username, last_activity) VALUES ($1, $2, $3)", c.Value, un, time.Now())
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
		tmpl.ExecuteTemplate(w, "catalogue.html", user)
		log.Println(user.ID + "signed up successfully")
		return
	}

	http.Redirect(w, r, "/signup.html", http.StatusSeeOther)
}

func login(w http.ResponseWriter, r *http.Request) {
	if alreadyLoggedIn(w, r) {
		http.Redirect(w, r, "/catalogue.html", http.StatusSeeOther)
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
		_, err = db.Exec("INSERT INTO sessions (uuid, username, last_activity) VALUES ($1, $2, $3)", c.Value, un, time.Now())
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		// redirect
		tmpl.ExecuteTemplate(w, "catalogue.html", user)
		log.Println(user.ID + " logged in successfully")
		return
	}
	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request) {
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

	http.Redirect(w, r, "/log-in", http.StatusSeeOther)
	log.Println("logged out" + c.Value)
}

func userinfo(w http.ResponseWriter, r *http.Request) {
	mem := getUser(w, r)
	fmt.Fprintln(w, mem.ID)
}

func submit(w http.ResponseWriter, r *http.Request) {
	mem := getUser(w, r)
	log.Println("turned into submit func")
	_, err := db.Exec("INSERT INTO submissions (username, problem, result, run_time, memory, submit_time, language) VALUES ($1, $2, $3, $4, $5, $6, $7)", 
	mem.ID, 1000, 0, 266, 728, time.Now(), 0)
	// result, run_time, memory, submit_time, language
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/status.html", http.StatusSeeOther)
}