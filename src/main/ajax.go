package main

import (
	"fmt"
	"log"
	"net/http"
)

func getUserInfo(w http.ResponseWriter, r *http.Request) {
	mem := getUser(w, r)
	fmt.Fprint(w, mem.ID)
}

func getPidCount(w http.ResponseWriter, r *http.Request) {
	var count int
	row := db.QueryRow("SELECT count(*) FROM problems")
	err := row.Scan(&count)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		log.Println("func getPid count pid error - ", err)
		return
	}

	fmt.Fprint(w, count)
}
