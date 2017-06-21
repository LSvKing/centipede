package centipede

import "net/http"

type Request struct {
	Url string

	// Responce type: html json jsonp text
	RespType string

	// GET POST
	Method string

	// POST data
	Postdata string

	Callback func(response *http.Response)

	//重试次数
	ReTry int
}
