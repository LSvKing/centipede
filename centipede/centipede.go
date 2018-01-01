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
	"time"

	"golang.org/x/time/rate"

	"centipede/common"
	"centipede/downloader"
	"centipede/items"
	"centipede/logs"
	"centipede/pipeline"
	"centipede/request"
	"centipede/resource_manage"
	"centipede/scheduler"
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

func GetCentipede() *Centipede {
	return &centipede
}

func Run() {
	defer func() {
		if p := recover(); p != nil {
			Log.WithField("trace", string(debug.Stack())).Errorf("蜘蛛异常错误 error: %v", p)
		}
	}()

	Log.Infoln("Centipede 开始运行")

	runtime.GOMAXPROCS(runtime.NumCPU())

	//var crawlerWait sync.WaitGroup

	ctx, _ := context.WithCancel(context.Background())

	Log.Debug(len(centipede.CrawlerJob))

	for crawlerChan := range centipede.CrawlerJob {

		Log.Debugln(crawlerChan.Option().Name)

		go func() {

			crawler := crawlerChan.CrawlerEr

			//crawlerWait.Add(1)

			go centipede.Pipeline.Run(crawler)

			defer func() {
				Log.Infoln(crawler.Option().Name, " Finish")
			}()

			Log.Infoln(crawlerChan.Option().Name + " 开始运行")

			mc := resource_manage.NewResourceManageChan(crawler.Option().Thread)

			crawler.Parse(crawler.Option().CallParams)

			transport := &http.Transport{
				DisableKeepAlives: true,
			}

			//funcMap := make(map[string]reflect.Value, 100)

			var limiter *rate.Limiter

			if crawler.Option().Limit > 0 {
				limiter = rate.NewLimiter(crawler.Option().Limit, int(crawler.Option().Limit))
			}

			for {
				if crawler.Option().Limit > 0 {
					limiter.Wait(ctx)
				}

				if centipede.Scheduler.Count() == 0 && mc.Has() == 0 {
					//crawlerWait.Done()
					break
				}

				mc.GetOne()

				req, _ := centipede.Scheduler.Poll()

				if req == nil {
					mc.FreeOne()
					continue
				}

				go func() {

					Log.Debug("start spider go")
					Log.Debug(req.GetUrl())

					defer func() {
						mc.FreeOne()
					}()

					if crawler.Option().Timeout > 0 {
						centipede.Downloader.Client.Timeout = crawler.Option().Timeout
					}

					if crawler.Option().DisableProxy {

						if len(crawler.Option().ProxyList) < 1 {
							Log.Error("代理池为空")
						}

						var rawProxy string

						if crawler.Option().ProxyFun != nil {
							rawProxy = crawler.Option().ProxyFun()
						} else {

							if len(crawler.Option().ProxyList) > 1 {
								//从代理持随机取代理
								rawProxy = crawler.Option().ProxyList[rand.Intn(len(crawler.Option().ProxyList))].ProxyURL

							} else {
								rawProxy = crawler.Option().ProxyList[0].ProxyURL
							}
						}

						proxy, err := url.Parse(rawProxy)

						if err != nil {
							Log.Error("代理转换格式失败")
						}

						transport.Proxy = http.ProxyURL(proxy)
						centipede.Downloader.Client.Transport = transport
					}

					if resp, err := centipede.Downloader.Download(req); err != nil {
						Log.WithField("type", "downloadReTry").WithField("Request", req).WithError(err).Error("下载重试")

						if req.ReTry < 4 {
							req.ReTry += 1

							centipede.Scheduler.Push(req)
						} else {
							Log.WithField("type", "downloadError").WithField("Request", req).WithError(err).Error("重试失败")
						}

					} else {
						defer func() {
							if p := recover(); p != nil {
								Log.WithField("trace", string(debug.Stack())).WithField("Request", req).Fatalf("蜘蛛异常错误 error: %v", p)
							}

						}()

						Log.Debugln("get finsh")

						//通过反射执行 回调函数
						params := make([]reflect.Value, 0)

						//取 resp 参数放入执行参数
						params = append(params, reflect.ValueOf(resp))

						//如果回调请求设置了 自定义参数 加入自定义参数到执行参数
						if req.CallParams != nil {
							params = append(params, reflect.ValueOf(req.CallParams))
						}

						//通过反射执行函数
						reflect.ValueOf(crawler).MethodByName(req.Callback).Call(params)
					}

					Log.Debug("start spider go end")
				}()

			}
		}()

		//crawlerWait.Wait()

	}

	<-centipede.CrawlerJob
}

func RunNew() {
	Log.Infoln("Centipede 开始运行")

	runtime.GOMAXPROCS(runtime.NumCPU())

	for crawlerChan := range centipede.CrawlerJob {

		ctx, _ := context.WithCancel(context.Background())
		limiter := rate.NewLimiter(crawlerChan.CrawlerEr.Option().Limit, 1024)

		go func() {
			limiter.Wait(ctx)

			req, err := centipede.Scheduler.Poll()

			fmt.Println(req)

			if err != nil {
				time.Sleep(1 * time.Second)
			}

			if req.ReTry > 3 {
				Log.Errorln("已到达最大重试次数,抛弃", req)
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

				Log.Debugln("get finsh")

				//通过反射执行 回调函数
				params := make([]reflect.Value, 0)

				//取 resp 参数放入执行参数
				params = append(params, reflect.ValueOf(resp))

				//如果回调请求设置了 自定义参数 加入自定义参数到执行参数
				if req.CallParams != nil {
					params = append(params, reflect.ValueOf(req.CallParams))
				}

				//通过反射执行函数
				reflect.ValueOf(crawlerChan.CrawlerEr).MethodByName(req.Callback).Call(params)
			}

		}()
	}

}

func AddCrawler(crawler items.CrawlerEr) {
	centipede.Crawlers[crawler.Option().Name] = crawler

	if crawler.Option().AutoRun {
		PushCrawler(crawler.Option().Name)
	}
}

// AddRequest 添加请求
func AddRequest(req *request.Request) {
	centipede.Scheduler.Push(req)
}

// AddCrawlerChan 加入爬虫到通道中
func AddCrawlerChan(crawlerName string) {
	centipede.CrawlerJob <- items.CrawlerChan{
		CrawlerEr: centipede.Crawlers[crawlerName],
	}
}

// AddData 添加数据
func AddData(data []items.Data, collection string) {
	defer func() {
		if err := recover(); err != nil {
			Log.Fatal(err)
		}
	}()

	t := time.Now()

	data = append(data, items.Data{
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

	req := request.NewRequest(url).AddPostParams(params)

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
		//Log.Debugln(c)
		centipede.CrawlerJob <- items.CrawlerChan{
			CrawlerEr: c,
		}
	} else {
		Log.Errorln(crawler, "爬虫脚本不存在")
	}
}

//同步请求
func Downloader(res *request.Request) (*http.Response, error) {
	return centipede.Downloader.Download(res)
}

func Close() {
	close(centipede.CrawlerJob)
}
