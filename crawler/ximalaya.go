package crawler

import (
	"centipede/centipede"
	"centipede/common"
	"centipede/items"
	"centipede/request"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Ximalaya struct {
	items.Crawler
}

type (
	xiMaList struct {
		Res      bool  `json:"res"`
		NextPage int   `json:"next_page"`
		SoundIds []int `json:"sound_ids"`
	}

	xiMaItem struct {
		ID       int
		PlayPath string `json:"play_path_64"`
		Duration int64  `json:"duration"`
		Title    string `json:"title"`
		NickName string `json:"nickname"`
		Intro    string `json:"intro"`
		CoverURL string `json:"cover_url"`
	}
)

func init() {
	//centipede.AddCrawler(&Ximalaya{
	//	items.Crawler{
	//		Name: "ximalaya",
	//	},
	//})
}

func (this *Ximalaya) Parse(params url.Values) {
	centipede.AddData([]items.Data{
		{
			Field: "t",
			Value: "T",
		},
	}, "ximalaya")
}

func (this *Ximalaya) Option() items.Crawler {
	return this.Crawler
}

func (this *Ximalaya) Pipeline(data items.DataRow) {
	//var items []Item
}

func (this *Ximalaya) ParseUrl(id string) {
	link := "http://www.ximalaya.com/album/" + id

	req := request.NewRequest(link)

	resp, err := centipede.Downloader(req)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		fmt.Println("资源不存在")
	}

	doc, _ := goquery.NewDocumentFromResponse(resp)

	resp.Body.Close()

	realLink := doc.Url.String()

	//if doc.Find(".mgr-5").Size() > 0 {
	mgr5 := doc.Find(".mgr-5").Text()

	pubdateArr := strings.Split(mgr5, ":")

	date := strings.TrimSpace(pubdateArr[1])

	t, err := time.Parse("2006-01-02", date)

	if err != nil {
		fmt.Println(err.Error())
	}

	image, _ := doc.Find(".albumface180 img").Attr("src")

	h, err := doc.Find(".personal_header .username").Html()

	if err != nil {
		fmt.Println(err)
	}

	re, _ := regexp.Compile(`(?Us)^(.*)\<`)

	fmt.Println(re.FindAllStringSubmatch(h, 1))
	nickname := strings.TrimSpace(re.FindAllStringSubmatch(h, 1)[0][1])

	rss := rss{
		Title: doc.Find(".detailContent_title h1").Text(),
		Author: []string{
			nickname,
		},
		Summary:     common.TrimHtml(doc.Find(".detailContent_intro article").Text()),
		Description: common.TrimHtml(doc.Find(".detailContent_intro article").Text()),
		Subtitle:    doc.Find(".detailContent_title h1").Text(),
		Version:     "2",
		Itunes:      "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Link:        link,
		Language:    "zh-cn",
		Image: Image{
			Href: image,
		},
		PubDate: t.Format(time.RFC1123),
		Owner: Owner{
			Name:  nickname,
			Email: "LSvKing@Gmail.com",
		},
	}

	if doc.Find(".mgr-5").Size() > 0 {
		mgr5 := doc.Find(".mgr-5").Text()

		pubdateArr := strings.Split(mgr5, ":")

		t, err := time.Parse("2006-01-02", strings.TrimSpace(pubdateArr[1]))

		if err != nil {
			fmt.Println(err.Error())
		}

		rss.PubDate = t.Format(time.RFC1123)
	}

	page := 1

	if doc.Find(".pagingBar .pagingBar_page").Length() > 0 {
		pageCount := doc.Find(".pagingBar .pagingBar_page").Last().Prev().Text()
		page, _ = strconv.Atoi(pageCount)
	}

	for i := 1; i <= page; i++ {
		u := realLink + "?page=" + strconv.Itoa(i)

		centipede.AddRequest(request.NewRequest(u).SetCallback("ParseItem"))
	}
}

func (this *Ximalaya) ParseList(response *http.Response) {
	docList, _ := goquery.NewDocumentFromResponse(response)

	docList.Find(".album_soundlist ul li").Each(func(i int, selection *goquery.Selection) {
		sound_id, _ := selection.Attr("sound_id")

		centipede.AddRequest(request.NewRequest("http://www.ximalaya.com/tracks/" + sound_id + ".json").SetCallback("ParseItem"))
	})
}

func (this *Ximalaya) ParseItem(response *http.Response) {
	var items []Item

	body, _ := ioutil.ReadAll(response.Body)

	var xiItem xiMaItem

	json.Unmarshal(body, &xiItem)

	//t, err := time.Parse("2006-01-02", strings.TrimSpace(selection.Find(".operate span").Text()))
	//
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//
	//pubDate := t.Format(time.RFC1123)

	items = append(items, Item{
		Title:    xiItem.Title,
		Subtitle: xiItem.Title,
		Author:   xiItem.NickName,
		//PubDate:  pubDate,
		Summary: xiItem.Intro,
		Guid: Guid{
			IsPermaLink: "true",
		},
		Image: Image{
			Href: xiItem.CoverURL,
		},
		Enclosure: Enclosure{
			Url:  xiItem.PlayPath,
			Type: "audio/mpeg",
		},
		Duration: xiItem.Duration,
	})
}
