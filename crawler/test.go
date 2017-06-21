package crawler

import (
	"douban_spider/centipede"
	"douban_spider/items"
	"fmt"
)

type TestCrawler struct {
	items.Crawler
}

func init() {
	centipede.AddCrawler(&TestCrawler{
		items.Crawler{
			Name:   "测试",
			Thread: 10,
			Limit:  1,
		},
	})
}

func (this *TestCrawler) Parse() {

}

func (this *TestCrawler) Option() items.Crawler {
	return this.Crawler
}

func (this *TestCrawler) Pipeline(data items.DataRow) {
	fmt.Println(data)
}
