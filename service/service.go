package service

import (
	"net/http"

	"centipede/centipede"
)

func AddHandler(w http.ResponseWriter, r *http.Request) {
	centipede.PushCrawler(r.FormValue("name"))
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))

	return
}

func CloseHandler(w http.ResponseWriter, r *http.Request) {
	centipede.Close()
}
