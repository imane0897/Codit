package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

func signup(w http.ResponseWriter, r *http.Request) {

	// LOG
	log.Println("---signup func---")
	log.Println(r.FormValue("username"))
	log.Println(r.FormValue("password"))
	log.Println(r.FormValue("password2"))

	if alreadyLoggedIn(w, r) {
		// LOG
		log.Println("already logged in")
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
		c.MaxAge = sessionLength
		http.SetCookie(w, c)
		dbSessions[c.Value] = session{un, time.Now()}

		// tore user in dbUsers
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
		http.Redirect(w, r, "/catalogue.html", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/signup.html", http.StatusSeeOther)
	log.Println("redirect to signup")
}
