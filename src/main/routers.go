package main

import (
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
	http.HandleFunc("/checkusername", checkusername)

	http.ListenAndServe(":9090", nil)
}

func signup(w http.ResponseWriter, r *http.Request) {

	// LOG
	log.Println("truned to signup func")
	log.Println(r.Method)
	log.Println(r.FormValue("username"))
	log.Println(r.FormValue("password"))
	log.Println(r.FormValue("password2"))
	// END

	if alreadyLoggedIn(w, r) {
		http.Redirect(w, r, "/catalogue.html", http.StatusSeeOther)
		return
	}
	var mem Member
	// process form submission
	if r.Method == http.MethodPost {
		// get form values
		un := r.FormValue("username")
		p := r.FormValue("password")
		// TODO: username taken?
		if _, ok := dbUsers[un]; ok {
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

		// TODO: store user in dbUsers
		bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		mem = Member{un, bs}
		dbUsers[un] = mem

		// redirect
		http.Redirect(w, r, "/catalogue.html", http.StatusSeeOther)
		return
	}
	showSessions() // for demonstration purposes
	http.Redirect(w, r, "/signup.html", http.StatusSeeOther)
	log.Println("redirect to signup")
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
		mem, ok := dbUsers[un]
		if !ok {
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}
		// does the entered password match the stored password?
		err := bcrypt.CompareHashAndPassword(mem.pwd, []byte(p))
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
		http.Redirect(w, r, "/catalogue.html", http.StatusSeeOther)
		return
	}
	showSessions() // for demonstration purposes
	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}


func checkusername(w http.ResponseWriter, r *http.Request) {
	log.Println("checked")
	log.Println(r.Method)
	log.Println(r.FormValue("username"))
	log.Println(r.FormValue("password"))
	log.Println(r.FormValue("password2"))
}