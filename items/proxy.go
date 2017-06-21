package items

type ProxyList []Proxy

type Proxy struct {
	ProxyURL string
	Weight   int
}
