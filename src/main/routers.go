package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"time"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("../../html")))
	http.HandleFunc("/sign-up", signup)
	http.HandleFunc("/log-in", login)
	http.HandleFunc("/status.html", showStatus)
	http.ListenAndServe(":9090", nil)
}

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

func showStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT * FROM submissions")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	subs := make([]ShowSubmission, 0)
	for rows.Next() {
		sub := ShowSubmission{}
		var st time.Time
		var lan int
		err := rows.Scan(&sub.RID, &sub.Username, &sub.Problem, &sub.Result, &sub.RunTime, &sub.Memory, &st, &lan)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		sub.SubmitTime = st.Format(time.RFC3339)
		switch lan {
		case 0:
			sub.Language = "C"
		case 1:
			sub.Language = "C++"
		case 2:
			sub.Language = "Java"
		}
		subs = append(subs, sub)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tmpl := template.Must(template.ParseFiles("status.html"))
	tmpl.Execute(w, subs)
}
