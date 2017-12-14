package crawler

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"centipede/centipede"
	"centipede/config"
	"centipede/items"
	"centipede/request"

	"io/ioutil"

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
			Thread:       50,
			Limit:        50,
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
		centipede.Log.Fatalln("db.Open(): %q\n", err)
	}

	defer sess.Close() // Remember to close the database session.

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

	fenLeiUrl := "http://m.ivoix.cn/fenlei"

	for i := 1; i <= 91; i++ {
		u := fenLeiUrl + strconv.Itoa(i)
		req := request.NewRequest(u).SetCallback("ParseFenUrl")

		centipede.AddRequest(req)
	}

	// req := request.NewRequest(u).SetCallback("ParseFenUrl")

	// centipede.AddRequest(req)
}

func (this *Ivoix) ParseFenUrl(response *http.Response) {

	doc, err := goquery.NewDocumentFromResponse(response)

	if err != nil {
		centipede.Log.Errorln("NewDocumentFromResponse", err)
	}

	pageNum := doc.Find(".pgsel option").Length()

	for i := 1; i < pageNum; i++ {
		req := request.NewRequest(response.Request.URL.String() + "p" + strconv.Itoa(i)).SetCallback("ParseFenList")
		centipede.AddRequest(req)
	}

	defer response.Body.Close()
}

func (this *Ivoix) ParseFenList(response *http.Response) {

	doc, err := goquery.NewDocumentFromResponse(response)

	if err != nil {
		centipede.Log.Errorln(err)
	}

	pagetit := doc.Find(".pagetit span").Eq(0).Text()

	//centipede.AddData([]items.Data{
	//	{
	//		Field: "name",
	//		Value: pagetit,
	//	},
	//}, "category")

	this.InsertMongo(map[string]interface{}{
		"name": pagetit,
	}, "category")

	doc.Find(".searchul li").Each(func(i int, selection *goquery.Selection) {
		_, ok := selection.Attr("data-role")

		if !ok {
			u := selection.Find("a").AttrOr("href", "null")
			req := request.NewRequest(siteUrl+u).SetCallback("ParseBookList").AddCallParam("category", pagetit)
			centipede.AddRequest(req)
		}
	})

	defer response.Body.Close()
}

func (this *Ivoix) ParseBookList(response *http.Response, params map[string]string) {

	doc, err := goquery.NewDocumentFromResponse(response)

	if err != nil {
		centipede.Log.Errorln("NewDocumentFromResponse : ", err)
	}

	pageNum := doc.Find(".pgsel option").Length()

	info := doc.Find("#bookinfo")

	author := info.Find("p").Eq(1).Text()

	if len(author) > 0 {
		if strings.Contains(author, "播音：") {
			author = strings.Replace(author, "播音：", "", -1)
		}
	}

	owner := info.Find("p").Eq(0).Text()

	if len(owner) > 0 {
		if strings.Contains(owner, "作者：") {
			owner = strings.Replace(owner, "作者：", "", -1)
		}
	}

	image := info.Find(".bookimg").AttrOr("src", "null")

	description := info.Find("p").Eq(5).Text()

	re := regexp.MustCompile(`book(\d+)`)
	bookStr := re.FindStringSubmatch(response.Request.URL.String())

	bookID := bookStr[1]

	p := map[string]interface{}{
		"bookId":      bookID,
		"category":    params["category"],
		"author":      author,
		"owner":       owner,
		"image":       image,
		"description": description,
	}

	this.InsertMongo(p, "book")

	//reqCover := request.NewRequest(image).SetCallback("DownloadCover").AddCallParam("bookId", bookID)
	//centipede.AddRequest(reqCover)

	for i := 1; i < pageNum; i++ {
		req := request.NewRequest(response.Request.URL.String()+"p"+strconv.Itoa(i)).SetCallback("ParseBook").AddCallParams(params).AddCallParam("bookId", bookID)
		centipede.AddRequest(req)
	}

	defer response.Body.Close()
}

func (this *Ivoix) ParseBook(response *http.Response, params map[string]string) {

	doc, err := goquery.NewDocumentFromResponse(response)

	if err != nil {
		centipede.Log.Errorln("NewDocumentFromResponse : ", err)
	}

	doc.Find("#sortedList li").Each(func(i int, selection *goquery.Selection) {
		aid := selection.Find("span").Eq(0).AttrOr("kv", "null")

		if aid != "null" {
			name := selection.Find("span").Eq(0).AttrOr("kt", "null")

			req := request.NewRequest(mp3Url).SetMethod("POST").
				AddHeader("Cookie", "safedog-flow-item=; lygusername=lsvking; userid=427591; ASPSESSIONIDQSTSTBCS=FKCGENOCPGGAJJCPNGEAFMIH; apwd=lsv324000; userid=427591; aname=lsvking; lyguserpwd=lsv324000; hisArt=%5B%7B%22title%22%3A%22%E5%A4%A9%E4%BD%93%E6%82%AC%E6%B5%AE%22%2C%22url%22%3A%22%2Fbook23549%22%7D%2C%7B%22title%22%3A%22%E5%87%AF%E5%8F%94%E8%A5%BF%E6%B8%B8%E8%AE%B0_1-5%E9%83%A8%E5%85%A8%E9%9B%86%22%2C%22url%22%3A%22%2Fbook23536%22%7D%2C%7B%22title%22%3A%22%E5%86%92%E6%AD%BB%E8%AE%B0%E5%BD%95%E7%A5%9E%E7%A7%98%E4%BA%8B%E4%BB%B64_%E9%9D%92%E9%9B%AA%E6%95%85%E4%BA%8B%22%2C%22url%22%3A%22%2Fbook22737%22%7D%2C%7B%22title%22%3A%22undefined%22%2C%22url%22%3A%22undefined%22%7D%2C%7B%22title%22%3A%22%E6%9D%91%E4%B8%8A%E6%98%A5%E6%A0%91_1Q84%22%2C%22url%22%3A%22%2Fbook23556%22%7D%2C%7B%22title%22%3A%22%E8%8B%8F%E9%BA%BB%E5%96%87%E5%A7%91%E4%BC%A0%22%2C%22url%22%3A%22%2Fbook23543%22%7D%2C%7B%22title%22%3A%22%E5%88%9D%E4%B8%AD%E7%94%9F%E5%BF%85%E8%83%8C%E5%8F%A4%E8%AF%97%E6%96%87%E6%A0%87%E5%87%86%E6%9C%97%E8%AF%B5%22%2C%22url%22%3A%22%2Fbook23513%22%7D%5D").
				AddPostParam("aid", aid).AddPostParam("uname", "lsvking").AddCallParam("aid", aid).AddCallParams(params).AddCallParam("name", name).
				SetCallback("ParseMp3")

			centipede.AddRequest(req)
		}

		if err != nil {
			centipede.Log.Errorln(err)
		}
	})

	defer response.Body.Close()
}

func (this *Ivoix) ParseMp3(response *http.Response, params map[string]string) {

	bodyBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		centipede.Log.Errorln("ReadAll : ", err)
	}

	filePath := string(bodyBytes)
	//
	//mUrl := downUrl + filePath

	//req := request.NewRequest(mUrl).SetCallback("DownloadMp3").AddCallParams(params)

	//req.CallParams = params

	//req.AddCallParam("filePath", filePath)

	//centipede.AddRequest(req)

	p := map[string]interface{}{
		"bookId":     params["bookId"],
		"name":       params["name"],
		"aid":        params["aid"],
		"path":       filePath,
		"updateTime": time.Now(),
	}

	this.InsertMongo(p, "audio")

	defer response.Body.Close()
}

func (this *Ivoix) DownloadMp3(response *http.Response, params map[string]string) {

	//client, err := oss.New("oss-cn-beijing-internal.aliyuncs.com", "j1wOLKZNFGcF9B0t", "7yKiMHweWpSJylcD02899v5eUH9nuG")
	//
	//if err != nil {
	//	centipede.Log.Errorln("oss client", err)
	//}
	//
	//bucket, err := client.Bucket("centipede")
	//
	//if err != nil {
	//	centipede.Log.Errorln("oss bucket", err)
	//}
	//
	//options := []oss.Option{
	//	oss.ContentType("audio/mp3"),
	//}
	//
	//err = bucket.PutObject(params["aid"], response.Body, options...)
	//
	//if err != nil {
	//	centipede.Log.Errorln("oss PutObject", err)
	//}

	p := map[string]interface{}{
		"bookId":     params["bookId"],
		"name":       params["name"],
		"aid":        params["aid"],
		"path":       params["filePath"],
		"updateTime": time.Now(),
	}

	this.InsertMongo(p, "audio")

	defer response.Body.Close()

}

func (this *Ivoix) DownloadCover(response *http.Response, params map[string]string) {
	client, err := oss.New("oss-cn-beijing-internal.aliyuncs.com", "j1wOLKZNFGcF9B0t", "7yKiMHweWpSJylcD02899v5eUH9nuG")

	if err != nil {
		centipede.Log.Errorln("oss client", err)
	}

	bucket, err := client.Bucket("centipede")

	if err != nil {
		centipede.Log.Errorln("oss bucket", err)
	}

	err = bucket.PutObject("cover/"+params["bookId"]+".jpg", response.Body)

	if err != nil {
		centipede.Log.Errorln("oss PutObject", "cover/"+params["bookId"]+".jpg", err)
	}

	defer response.Body.Close()
}

func (this *Ivoix) InsertMongo(data map[string]interface{}, collection string) {

	appConfig := config.Get()

	var settings = mongo.ConnectionURL{
		Host:     appConfig.Mongo.Host,     // server IP.
		Database: appConfig.Mongo.Database, // Database name.
	}

	settings.User = appConfig.Mongo.UserName
	settings.Password = appConfig.Mongo.PassWord

	sess, err := mongo.Open(settings)

	if err != nil {
		centipede.Log.Fatalln("db.Open(): %q\n", err)
	}

	defer sess.Close() // Remember to close the database session.

	c := sess.Collection(collection)

	r, err := c.Insert(data)

	if err != nil {
		centipede.Log.Errorln(err)
	}

	centipede.Log.Debugln(r)
}
