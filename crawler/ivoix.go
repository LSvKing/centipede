package crawler

import (
	"log"
	"net/http"
	"strconv"

	"centipede/centipede"
	"centipede/config"
	"centipede/items"
	"centipede/request"

	"io/ioutil"

	"strings"

	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"upper.io/db.v3/mongo"
)

type Ivoix struct {
	items.Crawler
}

var (
	siteUrl = "http://m.ivoix.cn"
	mp3Url  = "http://m.ivoix.cn/inc/audio.asp"
	downUrl = "http://125.46.58.23:88"
)

func init() {
	centipede.AddCrawler(&Ivoix{
		items.Crawler{
			Name:         "Ivoix",
			Thread:       5,
			Limit:        5,
			DisableProxy: false,
			Timeout:      time.Minute * 4,
			ProxyList: []items.Proxy{
				{
					ProxyURL: "http://127.0.0.1:1087",
				},
			},
			AutoRun: true,
		},
	})
}

func (this *Ivoix) Parse(params map[string]string) {
	this.ParseUrl()
}

func (this *Ivoix) Option() items.Crawler {
	return this.Crawler
}

func (this *Ivoix) Pipeline(data items.DataRow) {

	appConfig := config.Get()

	var settings = mongo.ConnectionURL{
		Host:     appConfig.Mongo.Host,     // server IP.
		Database: appConfig.Mongo.Database, // Database name.
	}

	settings.User = appConfig.Mongo.UserName
	settings.Password = appConfig.Mongo.PassWord

	sess, err := mongo.Open(settings)

	if err != nil {
		log.Fatalf("db.Open(): %q\n", err)
	}

	defer sess.Close() // Remember to close the database session.
	//
	//collections := make(Collections)
	//
	//for _, value := range dataCache {
	//
	//	if _, ok := collections[value.Collection]; !ok {
	//		collections[value.Collection] = sess.Collection(value.Collection)
	//	}
	//
	//	data := make(map[string]interface{})
	//
	//	for _, v := range value.Data {
	//		data[v.Field] = v.Value
	//	}
	//
	//	collections[value.Collection].Insert(data)
	//}

	collection := sess.Collection(data.Collection)

	mongoData := make(map[string]interface{})

	for _, v := range data.Data {
		mongoData[v.Field] = v.Value
	}

	r, err := collection.Insert(mongoData)

	if err != nil {
		centipede.Log.Errorln(err)
	}

	centipede.Log.Debugln(r)
}

func (this *Ivoix) ParseUrl() {
	centipede.Log.Debug("ivoix start")
	fenLeiUrl := "http://m.ivoix.cn/fenlei91"

	req := request.NewRequest(fenLeiUrl)

	resp, err := centipede.Downloader(req)

	defer func() {
		resp.Body.Close()
	}()

	if err != nil || resp.StatusCode == http.StatusNotFound {
		centipede.Log.Errorln("资源不存在")

		return
	}

	doc, err := goquery.NewDocumentFromResponse(resp)

	pageNum := doc.Find(".pgsel option").Length()

	for i := 1; i < pageNum; i++ {
		req := request.NewRequest(fenLeiUrl + "p" + strconv.Itoa(i)).SetCallback("ParseFenList")
		centipede.AddRequest(req)
	}
}

func (this *Ivoix) ParseFenList(response *http.Response) {

	doc, err := goquery.NewDocumentFromResponse(response)

	if err != nil {
		centipede.Log.Errorln(err)
	}

	pagetit := doc.Find(".pagetit span").Text()

	doc.Find(".searchul li").Each(func(i int, selection *goquery.Selection) {
		_, ok := selection.Attr("data-role")

		if !ok {
			u := selection.Find("a").AttrOr("href", "null")
			req := request.NewRequest(siteUrl+u).SetCallback("ParseBookList").AddCallParam("category", pagetit)
			centipede.AddRequest(req)
		}
	})
}

func (this *Ivoix) ParseBookList(response *http.Response, params map[string]string) {

	doc, err := goquery.NewDocumentFromResponse(response)

	if err != nil {
		centipede.Log.Errorln("NewDocumentFromResponse : ", err)
	}

	pageNum := doc.Find(".pgsel option").Length()

	for i := 1; i < pageNum; i++ {
		req := request.NewRequest(response.Request.URL.String()+"p"+strconv.Itoa(i)).SetCallback("ParseBook").AddCallParam("category", params["category"])
		centipede.AddRequest(req)
	}
}

func (this *Ivoix) ParseBook(response *http.Response, params map[string]string) {

	doc, err := goquery.NewDocumentFromResponse(response)

	if err != nil {
		centipede.Log.Errorln("NewDocumentFromResponse : ", err)
	}

	info := doc.Find("#bookinfo")

	author := info.Find("p").Eq(1).Text()
	owner := info.Find("p").Eq(0).Text()
	image := info.Find(".bookimg").AttrOr("src", "null")
	description := info.Find("p").Eq(5).Text()

	doc.Find("#sortedList li").Each(func(i int, selection *goquery.Selection) {
		aid := selection.Find("span").Eq(0).AttrOr("kv", "null")

		if aid != "null" {
			req := request.NewRequest(mp3Url).SetMethod("POST").
				AddHeader("Cookie", "safedog-flow-item=; lygusername=lsvking; userid=427591; ASPSESSIONIDQSTSTBCS=FKCGENOCPGGAJJCPNGEAFMIH; apwd=lsv324000; userid=427591; aname=lsvking; lyguserpwd=lsv324000; hisArt=%5B%7B%22title%22%3A%22%E5%A4%A9%E4%BD%93%E6%82%AC%E6%B5%AE%22%2C%22url%22%3A%22%2Fbook23549%22%7D%2C%7B%22title%22%3A%22%E5%87%AF%E5%8F%94%E8%A5%BF%E6%B8%B8%E8%AE%B0_1-5%E9%83%A8%E5%85%A8%E9%9B%86%22%2C%22url%22%3A%22%2Fbook23536%22%7D%2C%7B%22title%22%3A%22%E5%86%92%E6%AD%BB%E8%AE%B0%E5%BD%95%E7%A5%9E%E7%A7%98%E4%BA%8B%E4%BB%B64_%E9%9D%92%E9%9B%AA%E6%95%85%E4%BA%8B%22%2C%22url%22%3A%22%2Fbook22737%22%7D%2C%7B%22title%22%3A%22undefined%22%2C%22url%22%3A%22undefined%22%7D%2C%7B%22title%22%3A%22%E6%9D%91%E4%B8%8A%E6%98%A5%E6%A0%91_1Q84%22%2C%22url%22%3A%22%2Fbook23556%22%7D%2C%7B%22title%22%3A%22%E8%8B%8F%E9%BA%BB%E5%96%87%E5%A7%91%E4%BC%A0%22%2C%22url%22%3A%22%2Fbook23543%22%7D%2C%7B%22title%22%3A%22%E5%88%9D%E4%B8%AD%E7%94%9F%E5%BF%85%E8%83%8C%E5%8F%A4%E8%AF%97%E6%96%87%E6%A0%87%E5%87%86%E6%9C%97%E8%AF%B5%22%2C%22url%22%3A%22%2Fbook23513%22%7D%5D").
				AddPostParam("aid", aid).AddPostParam("uname", "lsvking").
				SetCallback("ParseMp3")

			name := selection.Find("span").Eq(0).AttrOr("kt", "null")

			p := map[string]string{
				"aid":         aid,
				"category":    params["category"],
				"name":        name,
				"author":      author,
				"owner":       owner,
				"image":       image,
				"description": description,
			}

			req.CallParams = p

			centipede.AddRequest(req)
		}

		if err != nil {
			centipede.Log.Errorln(err)
		}
	})
}

func (this *Ivoix) ParseMp3(response *http.Response, params map[string]string) {

	bodyBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		centipede.Log.Errorln("ReadAll : ", err)
	}

	bodyString := string(bodyBytes)

	mUrl := downUrl + bodyString

	req := request.NewRequest(mUrl).SetCallback("DownloadMp3")

	params["filePath"] = bodyString

	req.CallParams = params

	centipede.AddRequest(req)
}

func (this *Ivoix) DownloadMp3(response *http.Response, params map[string]string) {

	path := params["filePath"]

	sp := strings.Split(path, "/")

	path = strings.Replace(path, sp[len(sp)-1], "", -1)

	centipede.Log.Debugln(path, "====", sp[len(sp)-1])

	client, err := oss.New("oss-cn-beijing-internal.aliyuncs.com", "j1wOLKZNFGcF9B0t", "7yKiMHweWpSJylcD02899v5eUH9nuG")

	if err != nil {
		centipede.Log.Errorln("oss client", err)
	}

	bucket, err := client.Bucket("centipede")

	if err != nil {
		centipede.Log.Errorln("oss bucket", err)
	}

	err = bucket.PutObject(sp[len(sp)-1], response.Body)

	if err != nil {
		centipede.Log.Errorln("oss PutObject", err)
	}

	centipede.AddData([]items.Data{
		{
			Field: "name",
			Value: params["name"],
		},
		{
			Field: "category",
			Value: params["category"],
		},
		{
			Field: "author",
			Value: params["author"],
		},
		{
			Field: "owner",
			Value: params["owner"],
		},
		{
			Field: "image",
			Value: params["image"],
		},
		{
			Field: "description",
			Value: params["description"],
		},
		{
			Field: "path",
			Value: params["filePath"],
		},
	}, "ivoix")

}
