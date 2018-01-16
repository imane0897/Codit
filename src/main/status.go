package main

import (
	_ "github.com/lib/pq"
	"html/template"
	"net/http"
	"time"
)

func showStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT * FROM submissions ORDER BY rid DESC LIMIT 20")
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
		var res int

		err := rows.Scan(&sub.RID, &sub.Username, &sub.Problem, &res, &sub.RunTime, &sub.Memory, &st, &lan)
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
		switch res {
		case 0:
			sub.Result = "Pending"
		case 1:
			sub.Result = "Accept"
		case 2:
			sub.Result = "Wrong Answer"
		case 3:
			sub.Result = "Compile Error"
		case 4:
			sub.Result = "Runtime Error"
		case 5:
			sub.Result = "Time Limit Exceeded"
		case 6:
			sub.Result = "Memory Limit Exceeded"
		case 7:
			sub.Result = "Output Limit Exceeded"
		case 8:
			sub.Result = "Presentation Error"
		case 9:
			sub.Result = "System Error"
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
