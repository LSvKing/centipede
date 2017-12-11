//Package downloader 下载类
package downloader

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"centipede/config"
	"centipede/logs"
	"centipede/request"
)

var log = logs.New()

type Downloader struct {
	Client *http.Client
}

func New() *Downloader {
	appConfig := config.Get()

	var transport = &http.Transport{
		DisableKeepAlives: true,
	}

	if appConfig.HttpClient.ProxyDisable {
		proxyUrl := "http://" + appConfig.HttpClient.ProxyUser + ":" + appConfig.HttpClient.ProxyPassword + "@" + appConfig.HttpClient.ProxyHost + ":" + appConfig.HttpClient.ProxyPort
		proxy, _ := url.Parse(proxyUrl)
		transport.Proxy = http.ProxyURL(proxy)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(30 * time.Second),
	}

	client.Jar, _ = cookiejar.New(nil)

	return &Downloader{
		Client: client,
	}
}

func (downloader *Downloader) Download(req *request.Request) (*http.Response, error) {

	log.Debug("Download")

	httpReq, e := http.NewRequest(req.GetMethod(), req.GetUrl(), strings.NewReader(req.PostParams.Encode()))

	if e != nil {
		log.Errorln(e)
	}

	if req.Header != nil {
		httpReq.Header = req.Header
	}

	if req.GetMethod() == "POST" {
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if req.GetMethod() == "GET" {
		httpReq.URL.RawQuery = req.PostParams.Encode()
	}

	httpReq.Close = true

	client := downloader.Client

	resp, err := client.Do(httpReq)

	log.Debug("Download End")

	return resp, err
}
