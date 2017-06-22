package job

import (
	"github.com/LSvKing/cron"
	"github.com/LSvKing/centipede/centipede"
)

var (
	MainCron *cron.Cron
)

func Run(){
	MainCron = cron.New()
	MainCron.Start()
}

func Add(spec string,crawlerName string){
	MainCron.AddFunc(spec, func() {
		centipede.AddCrawlerChan(crawlerName)
	},"stock")

}