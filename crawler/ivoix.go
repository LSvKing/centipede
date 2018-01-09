package crawler

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"centipede/centipede"
	"centipede/config"
	"centipede/items"
	"centipede/request"

	"io/ioutil"

	"io"
	"os"

	"fmt"

	"github.com/JodeZer/mgop"
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/mgo.v2/bson"
)

type Ivoix struct {
	items.Crawler
}

var (
	siteUrl = "http://m.ivoix.cn"
	mp3Url  = "http://m.ivoix.cn/inc/audio.asp"
	downUrl = "http://125.46.58.23"

	appConfig = config.Get()

	mg mgop.SessionPool
)

type Book struct {
	Title       string `bson:"title"`
	BookID      string `bson:"bookId"`
	Category    string `bson:"category"`
	Author      string `bson:"author"`
	Owner       string `bson:"owner"`
	Image       string `bson:"image"`
	Description string `bson:"description"`
	Count       string `bson:"count"`
}

func init() {
	centipede.AddCrawler(&Ivoix{
		items.Crawler{
			Name:         "Ivoix",
			Thread:       10,
			Limit:        10,
			DisableProxy: true,
			Timeout:      time.Minute * 4,
			ProxyList: []items.Proxy{
				{
					ProxyURL: "http://H196AR4J9408XN6D:F766CDA5666E4627@http-dyn.abuyun.com:9020",
				},
			},
			ProxyFun: nil,
			AutoRun:  true,
		},
	})

	mg = getMongoSession()
}

func GetProxy() string {
	appConfig := config.Get()

	//locker := new(sync.Mutex)
	client := redis.NewClient(&redis.Options{
		Addr: appConfig.Redis.Host,
		DB:   1,
	})
ReGoto:
	if key, err := client.RandomKey().Result(); err != nil {
		goto ReGoto
	} else {
		if result, err := client.Get(key).Result(); err != nil {
			goto ReGoto
		} else {
			return string(result)
		}
	}
}

func (this *Ivoix) Parse(params map[string]string) {
	this.ParseUrl()
}

func (this *Ivoix) Option() items.Crawler {
	return this.Crawler
}

func (this *Ivoix) Pipeline(data items.DataRow) {
}

func (this *Ivoix) ParseUrl() {

	req := request.NewRequest("http://m.ivoix.cn/nav")
	response, err := centipede.Downloader(req)

	if err != nil {
		centipede.Log.Errorln("nav", err)
	}

	doc, err := goquery.NewDocumentFromResponse(response)

	if err != nil {
		centipede.Log.Errorln("NewDocumentFromResponse", err)
	}

	doc.Find(".bookList div").Each(func(i int, selection *goquery.Selection) {

		if dataRole, ok := selection.Attr("data-role"); ok {
			if dataRole == "collapsible" {

				selection.Find("li a").Each(func(i int, selection *goquery.Selection) {

					if href, exist := selection.Attr("href"); exist {
						fmt.Println(siteUrl + href)

						req := request.NewRequest(siteUrl + href).SetCallback("ParseFenUrl")
						centipede.AddRequest(req)

						upsetMongo(bson.M{"name": selection.Text()}, map[string]interface{}{
							"name": selection.Text(),
						}, "category")
					}
				})
			}
		}
	})

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

		req := request.NewRequest(response.Request.URL.String()).SetCallback("ParseBookList").AddCallParam("category", params["category"])
		req.ReTry += 1

		centipede.AddRequest(req)

		return
	}

	re := regexp.MustCompile(`book(\d+)`)
	bookStr := re.FindStringSubmatch(response.Request.URL.String())

	bookID := bookStr[1]

	pageNum := doc.Find(".pgsel option").Length()

	info := doc.Find("#bookinfo")

	count := info.Find("p").Eq(2).Text()

	if len(count) > 0 {
		if strings.Contains(count, "音频：") {
			count = strings.Replace(count, "音频：", "", -1)
		}
	}

	//存在且没有更新便跳过
	//if !checkBookUpdate(bookID, count) {
	//	return
	//}

	title := info.Find("h3").Text()

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

	p := map[string]interface{}{
		"title":       title,
		"bookId":      bookID,
		"category":    params["category"],
		"author":      author,
		"owner":       owner,
		"image":       image,
		"description": description,
		"count":       count,
	}

	upsetMongo(bson.M{"bookId": bookID}, p, "book")

	if image != "null" {
		reqCover := request.NewRequest(image).SetCallback("DownloadCover").AddCallParam("bookId", bookID)
		centipede.AddRequest(reqCover)
	}

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

		req := request.NewRequest(response.Request.URL.String()).SetCallback("ParseBook").AddCallParams(params).AddCallParam("bookId", params["bookId"])

		req.ReTry += 1
		centipede.AddRequest(req)

		return
	}

	doc.Find("#sortedList li").Each(func(i int, selection *goquery.Selection) {
		aid := selection.Find("span").Eq(0).AttrOr("kv", "null")

		if aid != "null" && checkAudioUpdate(aid) {
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

	if len(filePath) > 200 {
		centipede.Log.Errorln("error : (filePath) > 200", filePath)

		req := request.NewRequest(mp3Url).SetMethod("POST").
			AddHeader("Cookie", "ASPSESSIONIDCCQBCACB=COPINMMCPFMCOJIMCFPOMKLP; safedog-flow-item=; apwd=lsv324000; userid=427475; aname=13264180593").
			AddPostParam("aid", params["aid"]).AddPostParam("uname", "lsvking").AddCallParam("aid", params["aid"]).AddCallParams(params).AddCallParam("name", params["name"]).
			SetCallback("ParseMp3")

		req.ReTry += 1

		centipede.AddRequest(req)
	}
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

	upsetMongo(bson.M{"aid": params["aid"]}, p, "audio")

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
	//err = bucket.PutObject("cover/"+params["bookId"]+".jpg", response.Body)
	//
	//if err != nil {
	//	centipede.Log.Errorln("oss PutObject", "cover/"+params["bookId"]+".jpg", err)
	//}

	if response.Request.URL.String() == "null" || response.Request.URL.String() == "http://cover.ivoix.cn" {
		return
	}

	f, err := os.OpenFile(appConfig.FilePath+"/cover/"+params["bookId"]+".jpg", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		centipede.Log.Error("[文件下载]", err)
		return
	}

	//log.Debugf("存储时: %p",&file.Body,file.Body.ContentLength)
	_, err = io.Copy(f, response.Body)

	if err != nil {
		centipede.Log.Error("[文件下载]", err)
		return
	}

	defer response.Body.Close()
}

func (this *Ivoix) InsertMongo(data map[string]interface{}, collection string) {

	session := mg.AcquireSession()
	defer session.Release()

	db := session.DB(appConfig.Mongo.Database)

	c := db.C(collection)

	err := c.Insert(data)

	if err != nil {
		centipede.Log.Errorln(err)
	}
}

func upsetMongo(selector interface{}, data map[string]interface{}, collection string) {

	session := mg.AcquireSession()
	defer session.Release()

	db := session.DB(appConfig.Mongo.Database)

	c := db.C(collection)

	r, err := c.Upsert(selector, data)

	if err != nil {
		centipede.Log.Errorln(err)
	}

	centipede.Log.Debugln(r)
}

func checkBookUpdate(bookID string, count string) bool {
	session := mg.AcquireSession()
	defer session.Release()

	db := session.DB(appConfig.Mongo.Database)

	collection := db.C("book")

	var book Book

	if err := collection.Find(bson.M{"bookId": bookID}).One(&book); err == nil {
		if book.Count == count {
			return false
		}
	}

	return true
}

//检查Audio 是否存在 是否需要更新
func checkAudioUpdate(aid string) bool {
	session := mg.AcquireSession()
	defer session.Release()

	db := session.DB(appConfig.Mongo.Database)

	collection := db.C("book")

	if exist, _ := collection.Find(bson.M{"aid": aid}).Count(); exist > 0 {
		return false
	} else {
		return true
	}

}

func getMongoSession() mgop.SessionPool {

	mongoUrl := ""

	if appConfig.Mongo.UserName != "" {
		mongoUrl = "mongodb://" + appConfig.Mongo.UserName + ":" + appConfig.Mongo.PassWord + "@" + appConfig.Mongo.Host + ":" + appConfig.Mongo.Port + "/" + appConfig.Mongo.Database
	} else {
		mongoUrl = "mongodb://" + appConfig.Mongo.Host + ":" + appConfig.Mongo.Port + "/" + appConfig.Mongo.Database
	}
	//p, err := mgop.DialStrongPool(mongoUrl, 200)
	//session := p.AcquireSession()
	//defer session.Release()

	session, err := mgop.DialStrongPool(mongoUrl, 60)
	//defer session.Close()

	if err != nil {
		centipede.Log.Fatalln("db.Open(): %q\n", err)
	}

	return session
}
