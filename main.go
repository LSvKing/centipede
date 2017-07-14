package main

import (
	"net/http"

	"github.com/LSvKing/centipede/centipede"
	_ "github.com/LSvKing/centipede/crawler"
	"github.com/gorilla/mux"
	//"fmt"
	//"github.com/LSvKing/centipede/job"
	"github.com/LSvKing/centipede/service"
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
