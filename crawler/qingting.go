package crawler

import (
	"fmt"
	"net/http"

	"io/ioutil"

	"encoding/json"

	"strconv"

	"github.com/LSvKing/centipede/centipede"
	"github.com/LSvKing/centipede/items"
	"github.com/LSvKing/centipede/request"
)

type (
	QingTing struct {
		items.Crawler
	}

	rss struct {
		Link        string   `xml:"channel>link"`
		PubDate     string   `xml:"channel>pubDate"`
		Description string   `xml:"channel>description"`
		Itunes      string   `xml:"xmlns:itunes,attr"`
		Author      []string `xml:"channel>itunes:author"`
		Language    string   `xml:"channel>language"`
		Title       string   `xml:"channel>title"`
		Version     string   `xml:"version,attr"`
		Image       Image    `xml:"channel>itunes:image"`
		Summary     string   `xml:"channel>itunes:summary"`
		Subtitle    string   `xml:"channel>itunes:subtitle"`
		Owner       Owner    `xml:"channel>owner"`
		Item        []Item   `xml:"channel>item"`
	}

	Image struct {
		Href string `xml:"href,attr"`
	}

	Owner struct {
		Name  string `xml:"name"`
		Email string `xml:"email"`
	}

	Enclosure struct {
		Url  string `xml:"url,attr"`
		Type string `xml:"type,attr"`
	}

	Guid struct {
		Text        string `xml:",chardata"`
		IsPermaLink string `xml:"isPermaLink,attr"`
	}

	Item struct {
		Title     string    `xml:"title" json:"name"`
		Subtitle  string    `xml:"itunes:subtitle" json:"name"`
		Author    string    `xml:"itunes:author" json:"name"`
		PubDate   string    `xml:"pubDate"`
		Summary   string    `xml:"itunes:summary"`
		Guid      Guid      `xml:"guid"`
		Image     Image     `xml:"itunes:image"`
		Enclosure Enclosure `xml:"enclosure"`
		Duration  int64     `xml:"itunes:duration" json:"duration"`
	}

	//JSON struct
	Channels struct {
		Code   int
		Name   string
		ImgUrl string
		Data   struct {
			ID         int
			UpdateTime string
			PodCasters []struct {
				Desc   string
				ID     int
				UserId string
				Name   string
				ImgUrl string
			}
			Desc         string
			ProgramCount int
		}
	}

	ChannelItem struct {
		Data []struct {
			ID         int
			UpdateTime string
			Duration   int
			Desc       string
			FilePath   string
			Name       string
		}
	}
)

func init() {
	centipede.AddCrawler(&QingTing{
		items.Crawler{
			Name:         "蜻蜓",
			Thread:       10,
			Limit:        5,
			DisableProxy: false,
		},
	})
}

func (this *QingTing) Parse(params map[string]string) {
	this.parseUrl("220234")
}

func (this *QingTing) Option() items.Crawler {
	return this.Crawler
}

func (this *QingTing) Pipeline(data items.DataRow) {
	fmt.Println(data.Data)
}

func (this *QingTing) parseUrl(id string) {
	link := "http://i.qingting.fm/wapi/channels/" + id

	centipede.AddRequest(request.NewRequest(link).SetCallback("ParseList"))
}

func (this *QingTing) ParseList(response *http.Response) {
	defer func() {
		response.Body.Close()
	}()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		centipede.Log.Error(this.Name, " Error:", err)
	}

	var channels Channels
	err = json.Unmarshal(body, &channels)

	centipede.Log.Debugln(channels)

	if err != nil {
		centipede.Log.Error(this.Name, " Error:", err)
	}

	if channels.Code == 1 {
		centipede.Log.Error(this.Name, " Error:", "专辑不存在")

		return
	}

	pageCount := 1

	if channels.Data.ProgramCount > 30 {
		if channels.Data.ProgramCount%30 != 0 {
			pageCount = int(channels.Data.ProgramCount/30) + 1
		} else {
			pageCount = int(channels.Data.ProgramCount / 30)
		}
	}

	for i := 1; i <= pageCount; i++ {
		u := response.Request.URL.String() + "/programs/page/" + strconv.Itoa(i)
		centipede.AddRequest(request.NewRequest(u).SetCallback("ParseItem"))
	}

	//t, err := time.Parse("2006-01-02 15:04:05", channels.Data.UpdateTime)

	//if err != nil {
	//	centipede.Log.Error(this.Name, " Error:", err.Error())
	//}

}

func (this *QingTing) ParseItem(response *http.Response) {
	defer func() {
		response.Body.Close()
	}()

	body, err := ioutil.ReadAll(response.Body)

	centipede.Log.Debugln(response.Request.URL.String())

	if err != nil {
		centipede.Log.Error(this.Name, " Error:", err.Error())
	}

	var item ChannelItem

	json.Unmarshal(body, &item)

	centipede.AddData(
		items.Data{
			struct {
				Field string
				Value interface{}
			}{Field: "data", Value: item},
		}, "qingting")

}

func enclosureUrl(u string) string {
	return "http://od.qingting.fm/" + u
}

func guidUrl(u string) string {
	return "http://www.qingting.fm/#" + u
}
