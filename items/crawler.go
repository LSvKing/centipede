package items

import (
	"golang.org/x/time/rate"
)

type CrawlerEr interface {
	Parse(map[string]string)
	Option() Crawler
	Pipeline(data DataRow)
}

type Crawler struct {
	Name         string
	Limit        rate.Limit
	Thread       uint
	DisableProxy bool
	ProxyList    []Proxy
	AutoRun      bool
}

func (crawler Crawler) Register() Crawler {
	return crawler
}

type CrawlerChan struct {
	CrawlerEr
	Params map[string]string
}
