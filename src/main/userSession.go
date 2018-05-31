package main

import (
	"database/sql"
	// "net/http"
	"time"
	// "log"
	// "fmt"
	"log"
	"net/http"

	"github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
)

func getUser(ctx *fasthttp.RequestCtx) Member {
	// set new cookie
	var mem Member
	cookieValue := ctx.Request.Header.Cookie("session")
	if cookieValue != nil {
		// if the user exists already, get user
		var s Session
		row := db.QueryRow("SELECT * FROM sessions WHERE uuid = $1", cookieValue)
		err = row.Scan(&s.uuid, &s.username, &s.lastActivity)
		if err != sql.ErrNoRows {
			_, err = db.Exec("UPDATE sessions SET last_activity = $1 where uuid = $2", time.Now(), s.uuid)
			if err != nil {
				ctx.Error("Internal server error", http.StatusInternalServerError)
				log.Println("func getUser error: sql error")
			}
			row = db.QueryRow("SELECT * FROM members WHERE id = $1", s.username)
			err = row.Scan(&mem.ID, &mem.Password, &mem.Admin)
			if err != nil {
				ctx.Error("Internal server error", http.StatusInternalServerError)
				log.Println("func getUser error: cannot get user info")
			}
		}
	} else {
		sID, err := uuid.NewV4()
		if err != nil {
			ctx.Error("Internal server error", http.StatusInternalServerError)
			log.Println("func getUser error: unable to new session value")
		}
		var c fasthttp.Cookie
		c.SetKey("session")
		c.SetValue(sID.String())
		ctx.Response.Header.SetCookie(&c)
	}

	return mem
}

func alreadyLoggedIn(ctx *fasthttp.RequestCtx) bool {
	c := ctx.Request.Header.Cookie("session")

	// check session and user
	var s Session
	row := db.QueryRow("SELECT * FROM sessions WHERE uuid = $1", c)
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

// func isAdmin(ctx *fasthttp.RequestCtx) bool{
// 	mem := getUser(w, r)
// 	if mem.Admin == true {
// 		return true
// 	}
// 	return false
// }

// func cleanSessions() {
// 	// TODO:
// }
