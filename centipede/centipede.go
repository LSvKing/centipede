//Package centipede Centipede(蜈蚣)主类
package centipede

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/LSvKing/centipede/common"
	"github.com/LSvKing/centipede/downloader"
	"github.com/LSvKing/centipede/items"
	"github.com/LSvKing/centipede/logs"
	"github.com/LSvKing/centipede/pipeline"
	"github.com/LSvKing/centipede/request"
	"github.com/LSvKing/centipede/resource_manage"
	"github.com/LSvKing/centipede/scheduler"
)

type Centipede struct {
	Scheduler  *scheduler.Scheduler
	Downloader *downloader.Downloader
	Pipeline   *pipeline.Pipeline
	Limit      rate.Limit //每秒并发次数
	Crawlers   map[string]items.CrawlerEr
	CrawlerJob chan items.CrawlerChan
}

var (
	centipede = Centipede{
		Scheduler:  scheduler.New(),
		Downloader: downloader.New(),
		Pipeline:   pipeline.New(),
		Crawlers:   make(map[string]items.CrawlerEr),
		CrawlerJob: make(chan items.CrawlerChan, 1000),
	}

	Log = logs.New()
)

func Run1() {

	defer func() {
		if p := recover(); p != nil {
			Log.WithField("trace", string(debug.Stack())).Fatalf("蜘蛛异常错误 error: %v", p)
		}
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())

	var crawlerWait sync.WaitGroup

	fmt.Println("Spider Run")

	var limiter *rate.Limiter

	for _, crawler := range centipede.Crawlers {

		crawlerWait.Add(1)

		Log.Debug(crawler.Option().Name + " 开始运行")

		if crawler.Option().Limit > 0 {
			limiter = rate.NewLimiter(crawler.Option().Limit, 5)
		}

		ctx, _ := context.WithCancel(context.Background())

		mc := resource_manage.NewResourceManageChan(crawler.Option().Thread)

		transport := &http.Transport{
			DisableKeepAlives: true,
		}

		crawler.Parse(map[string]string{})

		funcMap := make(map[string]reflect.Value)

		centipede.Pipeline.Run(crawler)

		for {
			if crawler.Option().Limit > 0 {
				limiter.Wait(ctx)
			}

			if centipede.Scheduler.Count() == 0 && mc.Has() == 0 {
				crawlerWait.Done()
				fmt.Println("Spider Finfish")
				break
			}

			mc.GetOne()

			req := centipede.Scheduler.Poll()

			if req == nil {
				mc.FreeOne()
				continue
			}

			go func() {

				Log.Debug("start spider go")
				Log.Debug(req.GetUrl())

				defer func() {
					Log.Debug("free one")
					mc.FreeOne()
				}()

				if crawler.Option().DisableProxy {

					if len(crawler.Option().ProxyList) < 1 {
						Log.Error("代理持为空")
					}

					var rawProxy string

					if len(crawler.Option().ProxyList) > 1 {
						//从代理持随机取代理
						rawProxy = crawler.Option().ProxyList[rand.Intn(len(crawler.Option().ProxyList))].ProxyURL

					} else {
						rawProxy = crawler.Option().ProxyList[0].ProxyURL
					}

					proxy, err := url.Parse(rawProxy)

					if err != nil {
						Log.Error("代理转换格式失败")
					}

					transport.Proxy = http.ProxyURL(proxy)
					centipede.Downloader.Client.Transport = transport
				}

				if resp, err := centipede.Downloader.Download(req); err != nil {
					Log.WithField("type", "downloadReTry").WithError(err).Error("下载重试")

					if req.ReTry < 4 {
						req.ReTry += 1

						centipede.Scheduler.Push(req)
					} else {
						Log.WithField("type", "downloadError").WithError(err).Error("重试失败")
					}

				} else {
					defer func() {
						if p := recover(); p != nil {
							Log.WithField("trace", string(debug.Stack())).Fatalf("蜘蛛异常错误 error: %v", p)
						}

					}()

					Log.Debug("get finsh")

					params := make([]reflect.Value, 1)
					params[0] = reflect.ValueOf(resp)

					if callFunc, ok := funcMap[req.Callback]; ok {
						callFunc.Call(params)
					} else {
						r := reflect.ValueOf(crawler)

						if r.MethodByName(req.Callback).IsValid() == false {
							Log.Error(req)
							Log.Errorf(req.Callback + "回调函数不存在")
						} else {
							funcMap[req.Callback] = r.MethodByName(req.Callback)
							funcMap[req.Callback].Call(params)
						}

					}
				}

				Log.Debug("start spider go end")
			}()

		}

		crawlerWait.Wait()
	}

}

func Run() {
	defer func() {
		if p := recover(); p != nil {
			Log.WithField("trace", string(debug.Stack())).Fatalf("蜘蛛异常错误 error: %v", p)
		}
	}()

	Log.Debug("Centipede 开始运行")

	fmt.Println("Centipede 开始运行")

	runtime.GOMAXPROCS(runtime.NumCPU())

	var crawlerWait sync.WaitGroup

	ctx, _ := context.WithCancel(context.Background())

	Log.Debug(len(centipede.CrawlerJob))

	for crawlerChan := range centipede.CrawlerJob {

		Log.Debug(crawlerChan.Option().Name + " 开始运行")

		crawler := crawlerChan.CrawlerEr

		crawlerWait.Add(1)

		go centipede.Pipeline.Run(crawler)

		go func() {

			mc := resource_manage.NewResourceManageChan(crawler.Option().Thread)

			crawler.Parse(crawlerChan.Params)

			transport := &http.Transport{
				DisableKeepAlives: true,
			}

			funcMap := make(map[string]reflect.Value)

			var limiter *rate.Limiter

			if crawler.Option().Limit > 0 {
				limiter = rate.NewLimiter(crawler.Option().Limit, 5)
			}

			for {
				if crawler.Option().Limit > 0 {
					limiter.Wait(ctx)
				}

				if centipede.Scheduler.Count() == 0 && mc.Has() == 0 {
					crawlerWait.Done()
					fmt.Println("Spider Finfish")
					break
				}

				mc.GetOne()

				req := centipede.Scheduler.Poll()

				if req == nil {
					mc.FreeOne()
					continue
				}

				go func() {

					Log.Debug("start spider go")
					Log.Debug(req.GetUrl())

					defer func() {
						Log.Debug("free one")
						mc.FreeOne()
					}()

					if crawler.Option().DisableProxy {

						if len(crawler.Option().ProxyList) < 1 {
							Log.Error("代理持为空")
						}

						var rawProxy string

						if len(crawler.Option().ProxyList) > 1 {
							//从代理持随机取代理
							rawProxy = crawler.Option().ProxyList[rand.Intn(len(crawler.Option().ProxyList))].ProxyURL

						} else {
							rawProxy = crawler.Option().ProxyList[0].ProxyURL
						}

						proxy, err := url.Parse(rawProxy)

						if err != nil {
							Log.Error("代理转换格式失败")
						}

						transport.Proxy = http.ProxyURL(proxy)
						centipede.Downloader.Client.Transport = transport
					}

					if resp, err := centipede.Downloader.Download(req); err != nil {
						Log.WithField("type", "downloadReTry").WithError(err).Error("下载重试")

						if req.ReTry < 4 {
							req.ReTry += 1

							centipede.Scheduler.Push(req)
						} else {
							Log.WithField("type", "downloadError").WithError(err).Error("重试失败")
						}

					} else {
						defer func() {
							if p := recover(); p != nil {
								Log.WithField("trace", string(debug.Stack())).Fatalf("蜘蛛异常错误 error: %v", p)
							}

						}()

						Log.Debug("get finsh")

						params := make([]reflect.Value, 1)
						params[0] = reflect.ValueOf(resp)

						if callFunc, ok := funcMap[req.Callback]; ok {
							callFunc.Call(params)
						} else {
							r := reflect.ValueOf(crawler)

							if r.MethodByName(req.Callback).IsValid() == false {
								Log.Error(req)
								Log.Errorf(req.Callback + "回调函数不存在")
							} else {
								funcMap[req.Callback] = r.MethodByName(req.Callback)
								funcMap[req.Callback].Call(params)
							}

						}
					}

					Log.Debug("start spider go end")
				}()

			}
		}()

		crawlerWait.Wait()

	}

	<-centipede.CrawlerJob
}

func AddCrawler(crawler items.CrawlerEr) {
	centipede.Crawlers[crawler.Option().Name] = crawler
	PushCrawler(crawler.Option().Name)
}

// AddRequest 添加请求
func AddRequest(req *request.Request) {
	centipede.Scheduler.Push(req)
}

// AddCrawlerChan 加入爬虫到通道中
func AddCrawlerChan(crawlerName string) {
	centipede.CrawlerJob <- items.CrawlerChan{
		CrawlerEr: centipede.Crawlers[crawlerName],
		Params:    map[string]string{},
	}
}

// AddData 添加数据
func AddData(data items.Data, collection string) {
	defer func() {
		if err := recover(); err != nil {
			Log.Fatal(err)
		}
	}()

	t := time.Now()

	data = append(data, struct {
		Field string
		Value interface{}
	}{
		Field: "created",
		Value: t.Format("2006-01-02 15:04:05"),
	})

	centipede.Pipeline.DataChan <- items.DataRow{
		collection,
		data,
	}
}

// AddFile 添加文件
func AddFile(file items.File, path string) {
	centipede.Pipeline.FileChan <- items.FileRow{
		Path: path,
		File: file,
	}
}

// Login 登录获取Cookie
func Login(url string, params map[string]string) {

	req := request.NewRequest(url).AddParams(params)

	resp, _ := centipede.Downloader.Download(req)

	fmt.Println(resp.Body)

	centipede.Downloader.Client.Jar.SetCookies(resp.Request.URL, resp.Cookies())

	defer resp.Body.Close()
}

// SetCookie 设置Cookie
func SetCookie(rawUrl string, cookies []*http.Cookie) {
	u, err := url.Parse(rawUrl)

	if err != nil {
		Log.Error("Url无效")
	}

	centipede.Downloader.Client.Jar.SetCookies(u, cookies)
}

// SetCookieJSON 通过JSON设置Cookie
/**
[{
"domain": "domain.com",
"httpOnly": false,
"name": "_name",
"path": "/",
"secure": false,
"value": "google.com",
}]
*/
func SetCookieJSON(rawUrl string, cookie string) {

	var cookies []*http.Cookie
	u, err := url.Parse(rawUrl)

	if err != nil {
		Log.Error("Url无效")
	}

	err = json.Unmarshal([]byte(cookie), &cookies)

	if err != nil {
		Log.Error("JSON格式 Cookie 无效")
	}

	centipede.Downloader.Client.Jar.SetCookies(u, cookies)
}

// SetCookieString 通过字符串设置Cookie
/**
"mt=ci%3D-1_0; thw=cn; sec=5572dc7c40ce07d4e8c67e4879a; v=0;
*/
func SetCookieString(rawUrl string, cookie string) {
	u, err := url.Parse(rawUrl)

	if err != nil {
		Log.Error("Url无效")
	}

	cookies := common.SplitCookies(cookie)

	centipede.Downloader.Client.Jar.SetCookies(u, cookies)
}

func PushCrawler(crawler string) {
	if c, ok := centipede.Crawlers[crawler]; ok {
		Log.Debugln(c)
		centipede.CrawlerJob <- items.CrawlerChan{
			CrawlerEr: c,
			Params:    map[string]string{},
		}
	} else {
		Log.Errorln(crawler, "爬虫脚本不存在")
	}
}

func Close() {
	close(centipede.CrawlerJob)
}
