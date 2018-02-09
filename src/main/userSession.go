package main

import (
	"database/sql"
	"github.com/satori/go.uuid"
	"net/http"
	"time"
	"log"
)

func getUser(w http.ResponseWriter, r *http.Request) Member {
	// get cookie
	c, err := r.Cookie("session")
	if err != nil {
		sID, err := uuid.NewV4()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Println("func getUser error: cannot get cookie session")
		}
		c = &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
	}
	http.SetCookie(w, c)

	// if the user exists already, get user
	var s Session
	var mem Member
	row := db.QueryRow("SELECT * FROM sessions WHERE uuid = $1", c.Value)
	err = row.Scan(&s.uuid, &s.username, &s.lastActivity)
	if err != sql.ErrNoRows {
		_, err = db.Exec("UPDATE sessions SET last_activity = $1 where uuid = $2", time.Now(), s.uuid)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Println("func getUser error: sql error")
		}
		row = db.QueryRow("SELECT * FROM members WHERE id = $1", s.username)
		err = row.Scan(&mem.ID, &mem.Password)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Println("func getUser error: cannot get user info")
		}
	}
	return mem
}

func alreadyLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	c, err := r.Cookie("session")
	if err != nil {
		return false
	}

	// check session and user
	var s Session
	row := db.QueryRow("SELECT * FROM sessions WHERE uuid = $1", c.Value)
	err = row.Scan(&s.uuid, &s.username, &s.lastActivity)
	if err != sql.ErrNoRows {
		row = db.QueryRow("SELECT * FROM members WHERE id = $1", s.username)
		if row != nil {
			log.Println(s.username, "already logged in")
			return true
		}
	}
	return false
}

func cleanSessions() {
	// TODO:
}
