package service

import (
	"net/http"

	"github.com/LSvKing/centipede/centipede"
)

func AddHandler(w http.ResponseWriter, r *http.Request) {
	centipede.PushCrawler(r.FormValue("name"))
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))

	return
}
