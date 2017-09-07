package crawler

import (
	"fmt"
	"net/http"

	"github.com/LSvKing/centipede/centipede"
	"github.com/LSvKing/centipede/items"
	"github.com/LSvKing/centipede/request"
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
		},
	})
}

func (this *TestCrawler) Parse(params map[string]string) {
	centipede.AddData(
		items.Data{
			struct {
				Field string
				Value interface{}
			}{Field: "ttt", Value: "ssss"},
		}, "test")

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

	fmt.Println("Is Test Pipeline ", data)
}

func (this *TestCrawler) ParseItem(response *http.Response) {
	fmt.Println(response.Request.URL.String(), "FUCK")
}
