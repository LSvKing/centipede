package crawler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/LSvKing/centipede/centipede"
	"github.com/LSvKing/centipede/config"
	"github.com/LSvKing/centipede/items"
	"github.com/LSvKing/centipede/request"

	"github.com/PuerkitoBio/goquery"
	"upper.io/db.v3/mongo"
)

type DouBanMoviePic struct {
	items.Crawler
}

func init() {

	//proxyList:=[]items.Proxy{
	//	{
	//		ProxyURL:"",
	//	},
	//	{
	//		ProxyURL:"",
	//	},
	//	{
	//		ProxyURL:"",
	//	},
	//	{
	//		ProxyURL:"",
	//	},
	//	{
	//		ProxyURL:"",
	//	},
	//	{
	//		ProxyURL:"",
	//	},
	//	{
	//		ProxyURL:"",
	//	},
	//}
	//
	//
	//
	//centipede.AddCrawler(&DouBanMoviePic{
	//	items.Crawler{
	//		Name:"豆瓣图片",
	//		Thread:10,
	//		Limit:1,
	//		DisableProxy:false,
	//		ProxyList:proxyList,
	//	},
	//})
}

func (this *DouBanMoviePic) Parse() {
	//parseUrl()
}

func (this *DouBanMoviePic) Option() items.Crawler {
	return this.Crawler
}

func (this *DouBanMoviePic) Pipeline(data items.DataRow) {
	fmt.Println(data)
}

type Movie struct {
	Name string `title`
	DbId int    `dbId`
}

func parseUrl() {
	appConfig := config.Get()

	var settings = mongo.ConnectionURL{
		Host:     appConfig.Mongo.Host,     // server IP.
		Database: appConfig.Mongo.Database, // Database name.
	}

	mongo, err := mongo.Open(settings)

	if err != nil {
		centipede.Log.Fatalf("db.Open(): %q\n", err)
	}

	collection := mongo.Collection("movie")

	movies := collection.Find().Limit(1)

	var movie Movie

	urlPrefix := "https://movie.douban.com/subject/"

	for movies.Next(&movie) {
		url := urlPrefix + strconv.Itoa(movie.DbId)

		urlS := url + "/photos?type=S"

		urlR := url + "/photos?type=R"

		urlW := url + "/photos?type=W"

		reqS := request.NewRequest(urlS).SetCallback("ParseImageList")

		reqR := request.NewRequest(urlR).SetCallback("ParseImageList")

		reqW := request.NewRequest(urlW).SetCallback("ParseImageList")

		centipede.AddRequest(reqS)
		centipede.AddRequest(reqR)
		centipede.AddRequest(reqW)
	}
}

func (this *DouBanMoviePic) ParseImageList(response *http.Response) {
	doc, err := goquery.NewDocumentFromResponse(response)

	if err != nil {
		centipede.Log.Error("FromResponse失败", response.Request.URL.String())
	}

	url := response.Request.URL.String()

	total, exist := doc.Find(".thispage").Attr("data-total-page")

	if exist {
		totalNum, _ := strconv.Atoi(total)

		centipede.Log.Debugln("totalNum", totalNum)

		for i := 0; i < totalNum; i += 1 {
			start := strconv.Itoa(i * 40)
			req := &request.Request{
				Url:      url + "&start=" + start,
				Method:   "GET",
				Callback: "ParseImage",
			}

			centipede.AddRequest(req)
		}
	} else {
		req := &request.Request{
			Url:      url,
			Method:   "GET",
			Callback: "ParseImage",
		}

		centipede.AddRequest(req)
	}

}

func (this *DouBanMoviePic) ParseImage(response *http.Response) {
	doc, err := goquery.NewDocumentFromResponse(response)

	if err != nil {
		centipede.Log.Error("FromResponse失败", response.Request.URL.String())
	}

	doc.Find(".article").Find("ul li").Each(func(i int, s *goquery.Selection) {
		urlPrefix := "https://img3.doubanio.com/view/photo/raw/public/"

		id, err := s.Attr("data-id")

		if err {
			pId := "p" + id + ".jpg"
			url := urlPrefix + pId

			centipede.AddRequest(
				request.NewRequest(url).
					SetCallback("DownLoadImage").
					SetReferer("https://movie.douban.com/subject/" + id + "/all_photos"))
		}
	})
}

func (this *DouBanMoviePic) DownLoadImage(response *http.Response) {

	defer func() {
		if err := recover(); err != nil {
			centipede.Log.Errorf("err")
		}
	}()

	name := filepath.Base(response.Request.URL.String())

	centipede.AddFile(items.File{
		Response: response,
		FileName: name,
	}, "douban_image")
}
