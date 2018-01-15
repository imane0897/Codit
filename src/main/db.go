package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

var db *sql.DB

// Member has id in 3-10 char, pwd in 6-10 char.
type Member struct {
	id  string
	pwd []byte
}

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://root:password@localhost/codit?sslmode=disable")
	// db, err = sql.Open("postgres", "postgres://bond:password@localhost/bookstore?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	log.Println("Connected to database codit")
}

func _select() {
	db, err := sql.Open("postgres", "postgres://root:password@localhost/codit?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		panic(err)
	}
	log.Println("You connected to your database.")

	rows, err := db.Query("SELECT * FROM members;")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	members := make([]Member, 0)
	for rows.Next() {
		mem := Member{}
		err := rows.Scan(&mem.id, &mem.pwd) // order matters
		if err != nil {
			panic(err)
		}
		members = append(members, mem)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
}
