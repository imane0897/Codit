package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/", home)

	http.HandleFunc("/sign-up", signup)
	http.HandleFunc("/log-in", login)
	http.HandleFunc("/log-out", logout)
	http.HandleFunc("/status.html", showStatus)
	http.HandleFunc("/userinfo", userinfo)
	http.HandleFunc("/submit", submit)

	http.ListenAndServe(":9090", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	log.Println("turned into home func")
	http.FileServer(http.Dir("../../html"))
	tmpl.ExecuteTemplate(w, "index.html", nil)
	// if r.URL.Path != "/" {
	// 	w.WriteHeader(404)
    //     w.Write([]byte("<h1>404</h1>"))
	// } else {
	// 	// http.Handle("/", http.FileServer(http.Dir("../../html")))
	// 	// tmpl.ExecuteTemplate(w, "index.html", nil)
	// 	http.FileServer(http.Dir("../../html"))
    // }
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
		http.Redirect(w, r, "/catalogue", http.StatusSeeOther)
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

	http.Redirect(w, r, "/status.html", http.StatusSeeOther)
}
