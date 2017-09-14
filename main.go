package main

import (
	"net/http"

	"centipede/centipede"
	_ "centipede/crawler"
	"github.com/gorilla/mux"
	//"fmt"
	//"centipede/job"
	"centipede/service"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", service.TestHandler)
	r.HandleFunc("/add", service.AddHandler)
	r.HandleFunc("/close", service.CloseHandler)
	go http.ListenAndServe(":8000", r)

	centipede.Run()
	//job.Run()
	//fmt.Println()
}
