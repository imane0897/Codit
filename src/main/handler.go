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
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/bcrypt"
)

func homeHandler(ctx *fasthttp.RequestCtx) {
	mem := getUser(ctx)
	if mem.ID == "" {
		ctx.SetContentType("text/html; charset=utf-8")
		tmpl.ExecuteTemplate(ctx, "index.html", nil)
	} else {
		ctx.SetContentType("text/html; charset=utf-8")
		ctx.Redirect("/catalogue", http.StatusSeeOther)
	}
}

func signupHandler(ctx *fasthttp.RequestCtx) {
	if alreadyLoggedIn(ctx) {
		ctx.Redirect("/catalogue", fasthttp.StatusSeeOther)
		return
	}

	ctx.SetContentType("text/html; charset=utf-8")
	tmpl.ExecuteTemplate(ctx, "signup.html", nil)
}

func signupPostHandler(ctx *fasthttp.RequestCtx) {
	// get form values
	un := ctx.FormValue("username")
	p := ctx.FormValue("password")

	// username taken?
	row := db.QueryRow("SELECT * FROM members WHERE id = $1", un)
	user := Member{}
	err := row.Scan(&user.ID, &user.Password, &user.Admin)
	if err != sql.ErrNoRows {
		ctx.Error("Username already taken", fasthttp.StatusForbidden)
		return
	}

	// create session
	sID, _ := uuid.NewV4()
	var c fasthttp.Cookie
	c.SetKey("session")
	c.SetValue(sID.String())
	ctx.Response.Header.SetCookie(&c)

	_, err = db.Exec("INSERT INTO sessions (uuid, username, last_activity) VALUES ($1, $2, $3)",
		sID.String(), un, time.Now())
	if err != nil {
		ctx.Error("Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	// store user in dbUsers
	bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
	if err != nil {
		ctx.Error("Internal server error", http.StatusInternalServerError)
		return
	}
	_, err = db.Exec("INSERT INTO members (id, pwd) VALUES ($1, $2)", un, bs)
	if err != nil {
		ctx.Error("Internal server error", http.StatusInternalServerError)
		return
	}

	// redirect
	user.ID = string(un)
	ctx.Redirect("/catalogue", http.StatusSeeOther)
	log.Println(user.ID + " signed up successfully")
}

func loginHandler(ctx *fasthttp.RequestCtx) {
	if alreadyLoggedIn(ctx) {
		ctx.SetContentType("text/html; charset=utf-8")
		ctx.Redirect("/catalogue", http.StatusSeeOther)
		return
	}

	ctx.SetContentType("text/html; charset=utf-8")
	tmpl.ExecuteTemplate(ctx, "login.html", nil)
}

func loginPostHandler(ctx *fasthttp.RequestCtx) {
	un := ctx.FormValue("username")
	p := ctx.FormValue("password")

	// is there a username?
	row := db.QueryRow("SELECT * FROM members WHERE id = $1", un)
	user := Member{}
	err := row.Scan(&user.ID, &user.Password, &user.Admin)
	if err == sql.ErrNoRows {
		ctx.Error("Username and/or password do not match", http.StatusForbidden)
		return
	}

	// does the entered password match the stored password?
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(p))
	if err != nil {
		ctx.Error("Username and/or password do not match", http.StatusForbidden)
		return
	}

	// create session
	sID, _ := uuid.NewV4()
	var c fasthttp.Cookie
	c.SetKey("session")
	c.SetValue(sID.String())
	ctx.Response.Header.SetCookie(&c)
	_, err = db.Exec("INSERT INTO sessions (uuid, username, last_activity) VALUES ($1, $2, $3)",
		sID.String(), un, time.Now())

	if err != nil {
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// redirect
	log.Println(user.ID + " logged in successfully")
	ctx.Redirect("/catalogue", http.StatusSeeOther)
}

func logoutHandler(ctx *fasthttp.RequestCtx) {
	if !alreadyLoggedIn(ctx) {
		ctx.Redirect("/", http.StatusSeeOther)
		return
	}
	c := ctx.Request.Header.Cookie("session")

	// delete the session
	_, err := db.Exec("DELETE FROM sessions WHERE uuid = $1", c)
	if err != nil {
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// remove the cookie
	var expiredCookie fasthttp.Cookie
	expiredCookie.SetKey("session")
	expiredCookie.SetValue("")
	expiredCookie.SetExpire(fasthttp.CookieExpireDelete)
	ctx.Response.Header.SetCookie(&expiredCookie)

	ctx.Redirect("/login", http.StatusSeeOther)
}

func catalogueHandler(ctx *fasthttp.RequestCtx) {
	rows, err := db.Query("SELECT pid, title, level FROM problems ORDER BY pid ASC")
	if err != nil {
		ctx.Error(http.StatusText(500), 500)
		log.Println("func catalogueHandler cannot query problem catalogue in SQL -", err)
		return
	}
	defer rows.Close()

	pbs := make([]ProblemInfo, 0)
	mem := getUser(ctx)
	for rows.Next() {
		pbinfo := ProblemInfo{}

		// get Pid, Title, Level
		var level int
		err := rows.Scan(&pbinfo.Pid, &pbinfo.Title, &level)
		if err != nil {
			ctx.Error(http.StatusText(500), http.StatusInternalServerError)
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
			ctx.Error(http.StatusText(500), http.StatusInternalServerError)
			log.Println("func catalogueHandler cannot query problems -", err)
			return
		}
		row = db.QueryRow("SELECT count(*) FROM submissions WHERE problem = $1", pbinfo.Pid)
		err = row.Scan(&total)
		if err != nil {
			ctx.Error(http.StatusText(500), http.StatusInternalServerError)
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
		ctx.Error(http.StatusText(500), 500)
		return
	}

	ctx.SetContentType("text/html; charset=utf-8")
	tmpl.ExecuteTemplate(ctx, "catalogue.html", pbs)
}

func problemHandler(ctx *fasthttp.RequestCtx) {
	pid := ctx.FormValue("pid")
	if len(pid) == 0 {
		ctx.Error(http.StatusText(400), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM problems WHERE pid = $1", pid)

	pb := Problem{}
	var description, input, output, sampleInput, sampleOutput string
	err := row.Scan(&pb.Pid, &pb.Title, &description, &input, &output,
		&sampleInput, &sampleOutput, &pb.Level)
	switch {
	case err == sql.ErrNoRows:
		ctx.NotFound()
		log.Println("func problemHandler problem ", pid, " not found - ", err)
		return
	case err != nil:
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
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
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		log.Println("func problemHandler get total accepted error -", err)
		return
	}
	row = db.QueryRow("SELECT count(*) FROM submissions WHERE problem = $1", pb.Pid)
	err = row.Scan(&pb.Submissions)
	if err != nil {
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		log.Println("func problemHandler get total submissions error -", err)
		return
	}

	ctx.SetContentType("text/html; charset=utf-8")
	tmpl.ExecuteTemplate(ctx, "problem.html", pb)
}

func submitPostHandler(ctx *fasthttp.RequestCtx) {
	compiler := string(ctx.FormValue("compiler"))
	mem := getUser(ctx)
	if mem.ID == "" {
		ctx.Redirect("/login", http.StatusSeeOther)
		return
	}

	// write code to file
	atomic.AddUint64(&rid, 1)
	f, err := os.Create("../../filesystem/submissions/" + strconv.FormatUint(rid, 10) + "." + compiler)
	if err != nil {
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		log.Println("func submitHandler create file error -", err)
		return
	}
	defer f.Close()

	_, err = f.Write(ctx.FormValue("code"))
	if err != nil {
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		log.Println("func submitHandle write file error -", err)
		return
	}
	f.Sync()

	// get file type
	var ftype int
	switch compiler {
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
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		log.Println("func submitHandler insert submission info error -", err)
		return
	}

	ctx.Redirect("/status", http.StatusSeeOther)
}

func statusHandler(ctx *fasthttp.RequestCtx) {
	rows, err := db.Query("SELECT * FROM submissions ORDER BY rid DESC LIMIT 20")
	if err != nil {
		ctx.Error(http.StatusText(500), 500)
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
			ctx.Error(http.StatusText(500), 500)
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
		ctx.Error(http.StatusText(500), 500)
		return
	}

	ctx.SetContentType("text/html; charset=utf-8")
	tmpl.ExecuteTemplate(ctx, "status.html", subs)
}

func newPostHandler(ctx *fasthttp.RequestCtx) {
	pb := ProblemBytes{}
	pb.Pid = ctx.FormValue("pid")
	pb.Title = ctx.FormValue("title")
	pb.Level = ctx.FormValue("level")
	pb.Description = ctx.FormValue("description")
	pb.Input = ctx.FormValue("input")
	pb.Output = ctx.FormValue("output")
	pb.SampleInput = ctx.FormValue("sampleinput")
	pb.SampleOutput = ctx.FormValue("sampleoutput")

	_, err := db.Exec("INSERT INTO problems (pid, title, level, description, input, output, sample_input, sample_output) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		pb.Pid, pb.Title, pb.Level, pb.Description, pb.Input, pb.Output, pb.SampleInput, pb.SampleOutput)
	if err != nil {
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		log.Println("func editHandler db update error - ", err)
		return
	}

	ctx.Redirect("/dashboard", http.StatusSeeOther)
}

func editHandler(ctx *fasthttp.RequestCtx) {
	pid := ctx.FormValue("pid")
	if pid == nil {
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// get problem info from db
	row := db.QueryRow("SELECT * FROM problems WHERE pid = $1", pid)

	pb := ProblemBytes{}
	err := row.Scan(&pb.Pid, &pb.Title, &pb.Description, &pb.Input, &pb.Output,
		&pb.SampleInput, &pb.SampleOutput, &pb.Level)
	switch {
	case err == sql.ErrNoRows:
		ctx.NotFound()
		log.Println("func editHandler problem ", pid, " not found - ", err)
		return
	case err != nil:
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		log.Println("func editHandler problem ", pid, " query error - ", err)
		return
	}

	// encode to JSON
	response, err := json.Marshal(pb)
	if err != nil {
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		log.Println("func editHandler JSON marshal err - ", err)
		return
	}
	fmt.Fprintf(ctx, string(response))
}

func editPostHandler(ctx *fasthttp.RequestCtx) {
	pb := ProblemBytes{}
	pb.Pid = ctx.FormValue("pid")
	pb.Title = ctx.FormValue("title")
	pb.Level = ctx.FormValue("level")
	pb.Description = ctx.FormValue("description")
	pb.Input = ctx.FormValue("input")
	pb.Output = ctx.FormValue("output")
	pb.SampleInput = ctx.FormValue("sampleinput")
	pb.SampleOutput = ctx.FormValue("sampleoutput")

	_, err := db.Exec("UPDATE problems SET title = $1, level = $2, description = $3, input = $4, output = $5, sample_input = $6, sample_output = $7 WHERE pid = $8",
		pb.Title, pb.Level, pb.Description, pb.Input, pb.Output, pb.SampleInput, pb.SampleOutput, pb.Pid)
	if err != nil {
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		log.Println("func editHandler db update error - ", err)
		return
	}

	ctx.Redirect("/dashboard", http.StatusSeeOther)
}

func dashHandler(ctx *fasthttp.RequestCtx) {
	if isAdmin(ctx) == false {
		ctx.Redirect("/login", http.StatusSeeOther)
	}

	ctx.SetContentType("text/html; charset=utf-8")
	tmpl.ExecuteTemplate(ctx, "dashboard.html", nil)
}
