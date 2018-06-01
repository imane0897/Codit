package main

import (
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
	if(isAdmin(ctx)) {
		fmt.Fprint(ctx, "ture")
	} else {
		fmt.Fprint(ctx, "false")
	}
}
