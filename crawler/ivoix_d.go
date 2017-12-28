package crawler

import (
	"centipede/centipede"
	"centipede/config"
	"centipede/items"
	"net/http"
	"time"

	"io"
	"os"

	"centipede/request"

	"path/filepath"

	"github.com/JodeZer/mgop"
	"gopkg.in/mgo.v2/bson"
	"upper.io/db.v3/mongo"
)

type IvoixD struct {
	items.Crawler
}

type Audio struct {
	ID         bson.ObjectId `bson:"_id"`
	UpdateTime time.Time     `bson:"updateTime"`
	BookID     string        `bson:"bookId"`
	Name       string        `bson:"name"`
	Aid        string        `bson:"aid"`
	Path       string        `bson:"path"`
	IsDownload int           `bson:"isDownload"`
}

func init() {
	centipede.AddCrawler(&IvoixD{
		items.Crawler{
			Name:         "IvoixD",
			Thread:       10,
			Limit:        10,
			DisableProxy: true,
			Timeout:      time.Minute * 4,
			ProxyList: []items.Proxy{
				{
					ProxyURL: "http://HR03Y5983TE1C0MD:72DAB06BEF59368F@http-dyn.abuyun.com:9020",
				},
			},
			AutoRun: true,
		},
	})
}

var (
//siteUrl = "http://m.ivoix.cn"
//mp3Url  = "http://m.ivoix.cn/inc/audio.asp"
//downUrl = "http://125.46.58.23"
)

func (this *IvoixD) Parse(params map[string]string) {
	this.ParseUrl()
}

func (this *IvoixD) Option() items.Crawler {
	return this.Crawler
}

func (this *IvoixD) ParseUrl() {
	appConfig := config.Get()

	var settings = mongo.ConnectionURL{
		Host:     appConfig.Mongo.Host,     // server IP.
		Database: appConfig.Mongo.Database, // Database name.
	}

	settings.User = appConfig.Mongo.UserName
	settings.Password = appConfig.Mongo.PassWord

	p, err := mgop.DialStrongPool("mongodb://"+appConfig.Mongo.Host+":"+appConfig.Mongo.Post, 20)

	//"mongodb://"+appConfig.Mongo.UserName+":"+appConfig.Mongo.PassWord+"@"+appConfig.Mongo.Host+":"+appConfig.Mongo.Post
	session := p.AcquireSession()

	defer session.Release()

	if err != nil {
		centipede.Log.Fatalln("db.Open(): %q\n", err)
	}

	c := session.DB("centipede").C("audio")

	iter := c.Find(nil).Iter()

	if err != nil {
		centipede.Log.Fatalln("db.Open(): %q\n", err)
	}

	var audio Audio

	for iter.Next(&audio) {
		if len(audio.Path) < 150 {

			req := request.NewRequest(downUrl+audio.Path).AddCallParam("path", audio.Path).SetCallback("Download")

			//header := req.Header{
			//	"User-Agent": "Wget/1.18 (linux-gnu)",
			//}

			centipede.AddRequest(req)
		}

	}
}

func (this *IvoixD) Download(response *http.Response, params map[string]string) {
	appConfig := config.Get()

	d, err := os.Stat(appConfig.FilePath + filepath.Dir(params["path"]))

	if err != nil || d.IsDir() {
		if err := os.MkdirAll(appConfig.FilePath+filepath.Dir(params["path"]), 0777); err != nil {
			//logs.Log.Error(
			//	" *     Fail  [文件下载：%v | KEYIN：%v | 批次：%v]   %v [ERROR]  %v\n",
			//	self.Spider.GetName(), self.Spider.GetKeyin(), atomic.LoadUint64(&self.fileBatch), fileName, err,
			//)

			centipede.Log.Error("[创建目录]", err)
			return
		}
	}

	f, err := os.OpenFile(appConfig.FilePath+params["path"], os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
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

	defer func() {
		response.Body.Close()
		f.Close()
	}()
}

func (this *IvoixD) Pipeline(data items.DataRow) {

}
