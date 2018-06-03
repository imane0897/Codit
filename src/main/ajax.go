package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/valyala/fasthttp"
)

func getUserInfo(ctx *fasthttp.RequestCtx) {
	mem := getUser(ctx)
	fmt.Fprint(ctx, mem.ID)
}

func getPidCount(ctx *fasthttp.RequestCtx) {
	var count int
	row := db.QueryRow("SELECT count(*) FROM problems")
	err := row.Scan(&count)
	if err != nil {
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		log.Println("func getPid count pid error - ", err)
		return
	}

	fmt.Fprint(ctx, count)
}

func getAdmin(ctx *fasthttp.RequestCtx) {
	if isAdmin(ctx) {
		fmt.Fprint(ctx, "ture")
	} else {
		fmt.Fprint(ctx, "false")
	}
}

func getResult(ctx *fasthttp.RequestCtx) {
	sub := SubmissionResult{}
	row := db.QueryRow("SELECT result, run_time, memory FROM submissions ORDER BY rid DESC LIMIT 1")
	err := row.Scan(&sub.Result, &sub.RunTime, &sub.Memory)
	sub.RunTime = 15
	sub.Memory = 248
	if err != nil {
		ctx.Error(http.StatusText(500), 500)
		log.Println("func getResult scan submission erro -", err)
		return
	}

	// encode to JSON
	response, err := json.Marshal(sub)
	if err != nil {
		ctx.Error(http.StatusText(500), http.StatusInternalServerError)
		log.Println("func getResult JSON marshal err - ", err)
		return
	}
	fmt.Fprint(ctx, string(response))
}
