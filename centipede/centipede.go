//Package centipede Centipede(蜈蚣)主类
package centipede

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

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
	CrawlerJob chan items.CrawlerEr
}

var (
	centipede = Centipede{
		Scheduler:  scheduler.New(),
		Downloader: downloader.New(),
		Pipeline:   pipeline.New(),
		Crawlers:make(map[string]items.CrawlerEr),
		CrawlerJob : make(chan items.CrawlerEr),
	}

	log = logs.New()
)

func Run() {

	defer func() {
		if p := recover(); p != nil {
			log.WithField("trace", string(debug.Stack())).Fatalf("蜘蛛异常错误 error: %v", p)
		}
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())

	var crawlerWait sync.WaitGroup

	fmt.Println("Spider Run")

	var limiter *rate.Limiter

	for _, crawler := range centipede.Crawlers {

		crawlerWait.Add(1)

		log.Debug(crawler.Option().Name + " 开始运行")

		if crawler.Option().Limit > 0 {
			limiter = rate.NewLimiter(crawler.Option().Limit, 5)
		}

		ctx, _ := context.WithCancel(context.Background())

		mc := resource_manage.NewResourceManageChan(crawler.Option().Thread)

		transport := &http.Transport{
			DisableKeepAlives: true,
		}

		crawler.Parse()

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

				log.Debug("start spider go")
				log.Debug(req.GetUrl())

				defer func() {
					log.Debug("free one")
					mc.FreeOne()
				}()

				if crawler.Option().DisableProxy {

					if len(crawler.Option().ProxyList) < 1 {
						log.Error("代理持为空")
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
						log.Error("代理转换格式失败")
					}

					transport.Proxy = http.ProxyURL(proxy)
					centipede.Downloader.Client.Transport = transport
				}

				if resp, err := centipede.Downloader.Download(req); err != nil {
					log.WithField("type", "downloadReTry").WithError(err).Error("下载重试")

					if req.ReTry < 4 {
						req.ReTry += 1

						centipede.Scheduler.Push(req)
					} else {
						log.WithField("type", "downloadError").WithError(err).Error("重试失败")
					}

				} else {
					defer func() {
						if p := recover(); p != nil {
							log.WithField("trace", string(debug.Stack())).Fatalf("蜘蛛异常错误 error: %v", p)
						}

					}()

					log.Debug("get finsh")

					params := make([]reflect.Value, 1)
					params[0] = reflect.ValueOf(resp)

					if callFunc, ok := funcMap[req.Callback]; ok {
						callFunc.Call(params)
					} else {
						r := reflect.ValueOf(crawler)

						if r.MethodByName(req.Callback).IsValid() == false {
							log.Error(req)
							log.Errorf(req.Callback + "回调函数不存在")
						} else {
							funcMap[req.Callback] = r.MethodByName(req.Callback)
							funcMap[req.Callback].Call(params)
						}

					}
				}

				log.Debug("start spider go end")
			}()

		}

		crawlerWait.Wait()
	}

}

func (centipede Centipede) Run(){
	//c := centipede.Crawlers[0]
	//
	//centipede.CrawlerJob <- c
}

func AddCrawler(crawler items.CrawlerEr) {
	centipede.Crawlers[crawler.Option().Name] = crawler
}

// AddRequest 添加请求
func AddRequest(req *request.Request) {
	centipede.Scheduler.Push(req)
}

// AddCrawlerChan 加入爬虫到通道中
func AddCrawlerChan(crawlerName string){
	centipede.CrawlerJob <- centipede.Crawlers[crawlerName]
}

// AddData 添加数据
func AddData(data items.Data, collection string) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
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
		log.Error("Url无效")
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
		log.Error("Url无效")
	}

	err = json.Unmarshal([]byte(cookie), &cookies)

	if err != nil {
		log.Error("JSON格式 Cookie 无效")
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
		log.Error("Url无效")
	}

	cookies := common.SplitCookies(cookie)

	centipede.Downloader.Client.Jar.SetCookies(u, cookies)
}
