package service

import (
	"encoding/json"
	"net/http"

	"centipede/centipede"
	"centipede/items"
	"fmt"
)

//添加蜘蛛到执行队列
func AddHandler(w http.ResponseWriter, r *http.Request) {
	centipede.PushCrawler(r.FormValue("name"))
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))

	return
}

func JobList(w http.ResponseWriter, r *http.Request) {

	if len(centipede.GetCentipede().Crawlers) > 0 {
		i := 1

		for _, c := range centipede.GetCentipede().Crawlers {
			fmt.Print(i)
			fmt.Print(" : ")
			fmt.Printf("%+v\n", c.Option())
		}
	}
}

func CrawlerList(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if len(centipede.GetCentipede().Crawlers) > 0 {
		i := 1

		var cList []items.Crawler

		for _, c := range centipede.GetCentipede().Crawlers {
			fmt.Print(i)
			fmt.Print(" : ")
			fmt.Printf("%+v\n", c.Option())

			cList = append(cList, c.Option())
		}

		if err := json.NewEncoder(w).Encode(cList); err != nil {
			panic(err)
		}
	} else {
		m := items.Message{
			Message: "Null",
		}

		if err := json.NewEncoder(w).Encode(m); err != nil {
			panic(err)
		}
	}
}

//结束抓取程序
func CloseHandler(w http.ResponseWriter, r *http.Request) {
	centipede.Close()
}
