package main

import (
	"net/http"

	"centipede/centipede"
	_ "centipede/crawler"
	"centipede/service"

	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()

	//r.HandleFunc("/", service.TestHandler)
	r.HandleFunc("/add", service.AddHandler)
	r.HandleFunc("/close", service.CloseHandler)
	r.HandleFunc("/crawlers", service.CrawlerList)
	go http.ListenAndServe(":8000", r)

	centipede.Run()
	//job.Run()
	//fmt.Println()
}
