package crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"centipede/centipede"
	"centipede/config"
	"centipede/items"
	"centipede/request"

	"github.com/davecgh/go-spew/spew"
	"upper.io/db.v3/mongo"
)

type Stock struct {
	items.Crawler
}

var AppKey = "da1f72d51bc9de9bee4b64694c585b0f"

type StockList struct {
	Error_code int
	Reason     string
	Result     struct {
		TotalCount string
		Page       string
		Num        string
		Data       []struct {
			Symbol        string
			Name          string
			Trade         string
			Pricechange   string
			Changepercent string
			Buy           string
			Sell          string
			Settlement    string
			Open          string
			High          string
			Low           string
			Volume        int
			Amount        int
			Code          string
			Ticktime      string
		}
	}
}

func init() {
	//stock := Stock{
	//	items.Crawler{
	//		Name:         "股票",
	//		Limit:        5,
	//		Thread:       10,
	//		DisableProxy: false,
	//	},
	//}
	//
	//centipede.AddCrawler(&stock)
}

func (stock *Stock) Parse() {
	req := request.NewRequest("http://web.juhe.cn:8080/finance/stock/shall").SetCallback("StockCodeList").AddPostParam("key", AppKey).SetMethod("POST")

	centipede.AddRequest(req)
}

func (stock *Stock) Option() items.Crawler {
	return stock.Crawler
}

func (stock *Stock) Pipeline(data items.DataRow) {
	fmt.Println(data)
}

func (stock *Stock) StockCodeList(response *http.Response) {
	responseData, err := ioutil.ReadAll(response.Body)

	if err != nil {
		centipede.Log.Fatal(err)
	}

	var r StockList

	err = json.Unmarshal(responseData, &r)

	if err != nil {
		fmt.Printf("format err:%s\n", err.Error())
		return
	}

	if r.Error_code == 0 {
		totalCount, _ := strconv.Atoi(r.Result.TotalCount)

		totalPage := int(totalCount/20 + 1)

		spew.Dump(totalPage)

		u := response.Request.URL.String()

		for i := 1; i <= totalPage; i++ {
			req := request.NewRequest(u).AddPostParams(
				map[string]string{
					"key":  AppKey,
					"page": strconv.Itoa(i),
				},
			).SetCallback("GetStock")

			centipede.AddRequest(req)
		}
	}
}

func (stock *Stock) GetStock(response *http.Response) {
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		centipede.Log.Fatal(err)
	}

	var r StockList

	json.Unmarshal(responseData, &r)

	appConfig := config.Get()

	var settings = mongo.ConnectionURL{
		Host:     appConfig.Mongo.Host, // server IP.
		Database: "stock",              // Database name.
	}

	mongo, err := mongo.Open(settings)

	if err != nil {
		centipede.Log.Fatalf("db.Open(): %q\n", err)
	}

	collection := mongo.Collection("stock_list2")

	if r.Error_code == 0 {

		for _, item := range r.Result.Data {
			collection.Insert(item)
		}
	}
}
