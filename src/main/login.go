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

func login(w http.ResponseWriter, r *http.Request) {
	if alreadyLoggedIn(w, r) {
		// LOG
		log.Println("already logged in")
		http.Redirect(w, r, "/catalogue.html", http.StatusSeeOther)
		return
	}
	// LOG
	log.Println(r.FormValue("username"))
	log.Println(r.FormValue("password"))

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
		c.MaxAge = sessionLength
		http.SetCookie(w, c)
		dbSessions[c.Value] = session{un, time.Now()}

		// redirect
		http.Redirect(w, r, "/catalogue.html", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
	log.Println("redirect to login")
}
