//Package request Request请求类
package request

import (
	"net/http"
	urllib "net/url"
)

type Request struct {
	Url string

	// Responce type: html json jsonp text
	RespType string

	// GET POST
	Method string

	// POST data
	Params urllib.Values

	// 回调函数名称
	Callback string

	//重试次数
	ReTry int

	Header http.Header

	Cookie http.Cookie
}

func NewRequest(url string) *Request {

	request := Request{
		Url:    url,
		Header: http.Header{},
		Method: "GET",
		Params: urllib.Values{},
	}

	return &request
}

func (this *Request) SetCallback(callback string) *Request {
	this.Callback = callback
	return this
}

func (this *Request) SetReferer(referer string) *Request {
	this.Header.Add("Referer", referer)
	return this
}

func (this *Request) GetUrl() string {
	return this.Url
}

func (this *Request) GetMethod() string {

	if this.Method != "" {
		return this.Method
	} else {
		return "GET"
	}

}

func (this *Request) GetParams() string {
	return this.Params.Encode()
}

func (this *Request) AddParam(key string, value string) *Request {
	this.Params.Add(key, value)

	return this
}

func (this *Request) AddParams(params map[string]string) *Request {
	for k, v := range params {
		this.Params.Add(k, v)
	}

	return this
}

func (this *Request) SetParam(key string, value string) *Request {
	this.Params.Set(key, value)
	return this
}

func (this *Request) SetUrl(url string) *Request {
	this.Url = url
	return this
}

func (this *Request) SetMethod(method string) *Request {
	this.Method = method
	return this
}

func (this *Request) AddHeader(key string, value string) *Request {
	this.Header.Add(key, value)
	return this
}

func (this *Request) SetHeader(key string, value string) *Request {
	this.Header.Set(key, value)
	return this
}

func (this *Request) SetCookie(cookie string) *Request {
	return this
}

func (this *Request) SetUserAgent(userAgent string) *Request {
	this.Header.Add("User-Agent", userAgent)
	return this
}

func (this *Request) SetContentType(contentType string) *Request {
	this.Header.Add("Content-Type", contentType)
	return this
}
