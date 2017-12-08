package items

import (
	"time"

	"golang.org/x/time/rate"
)

type CrawlerEr interface {
	Parse(map[string]string)
	Option() Crawler
	Pipeline(data DataRow)
}

type Crawler struct {
	Name         string     `json:"name"`
	Limit        rate.Limit `json:"limit"`
	Thread       uint       `json:"thread"`
	Timeout      time.Duration
	DisableProxy bool              `json:"disableProxy"`
	ProxyList    []Proxy           `json:"proxyList"`
	AutoRun      bool              `json:"autoRun"`
	CallParams   map[string]string `json:"params"`
}

func (crawler Crawler) Register() Crawler {
	return crawler
}

type CrawlerChan struct {
	CrawlerEr
}
