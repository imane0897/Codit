package main

import (
	_ "github.com/lib/pq"
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("../../html")))
	http.HandleFunc("/sign-up", signup)
	http.HandleFunc("/log-in", login)
	http.HandleFunc("/status.html", showStatus)
	http.ListenAndServe(":9090", nil)
}
