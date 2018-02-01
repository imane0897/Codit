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
	"time"
)

func init() {
	tmpl = template.Must(template.ParseGlob("../../view/*.html"))

	// connect to database
	var err error
	db, err = sql.Open("postgres", "postgres://root:password@localhost/codit?sslmode=disable")
	if err != nil {
		log.Println("database open error: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Println("database connect error: ", err)
	}
	log.Println("Connected to the database")
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("../../view/")))

	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/catalogue", catalogueHandler)
	http.HandleFunc("/status", showStatusHandler)
	http.HandleFunc("/userinfo", userInfoHandler)
	http.HandleFunc("/submit", submitHandler)

	http.ListenAndServe(":9090", nil)
}

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
		_, err = db.Exec("INSERT INTO sessions (uuid, username, last_activity) VALUES ($1, $2, $3)", c.Value, un, time.Now())
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
	tmpl.ExecuteTemplate(w, "catalogue.html", nil)
}

func userInfoHandler(w http.ResponseWriter, r *http.Request) {
	mem := getUser(w, r)
	fmt.Fprintln(w, mem.ID)
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
