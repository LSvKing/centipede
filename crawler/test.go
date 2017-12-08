package crawler

import (
	"fmt"
	"net/http"

	"centipede/centipede"
	"centipede/items"
	"centipede/request"
)

type TestCrawler struct {
	items.Crawler
}

func init() {
	centipede.AddCrawler(&TestCrawler{
		items.Crawler{
			Name:         "Test",
			Thread:       10,
			Limit:        5,
			DisableProxy: false,
			CallParams:   map[string]string{},
			AutoRun:      false,
		},
	})
}

func (this *TestCrawler) Parse(params map[string]string) {

	centipede.AddRequest(request.NewRequest("http://baidu.com").SetCallback("ParseItem"))
	centipede.AddRequest(request.NewRequest("http://baidu.com").SetCallback("ParseItem"))
	centipede.AddRequest(request.NewRequest("http://baidu.com").SetCallback("ParseItem"))
	centipede.AddRequest(request.NewRequest("http://baidu.com").SetCallback("ParseItem"))
	centipede.AddRequest(request.NewRequest("http://baidu.com").SetCallback("ParseItem"))
	centipede.AddRequest(request.NewRequest("http://baidu.com").SetCallback("ParseItem"))
	centipede.AddRequest(request.NewRequest("http://baidu.com").SetCallback("ParseItem"))
	centipede.AddRequest(request.NewRequest("http://baidu.com").SetCallback("ParseItem"))
	centipede.AddRequest(request.NewRequest("http://baidu.com").SetCallback("ParseItem"))
	centipede.AddRequest(request.NewRequest("http://baidu.com").SetCallback("ParseItem"))

}

func (this *TestCrawler) Option() items.Crawler {
	return this.Crawler
}

func (this *TestCrawler) Pipeline(data items.DataRow) {
	//fmt.Println(config.Get().Mongo.Host)
	fmt.Println("Is Test Pipeline ", data)
}

func (this *TestCrawler) ParseItem(response *http.Response) {
	//fmt.Println(response.Request.URL.String(), "FUCK")

	centipede.AddData([]items.Data{
		{Field: "test", Value: "TEST"}, {Field: "p1", Value: 'p'},
	}, "test")
}

func (this *TestCrawler) Te(s string, b string) {
	fmt.Println("i'm te ", s, b)
}
