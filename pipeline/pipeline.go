package pipeline

import (
	"io"
	"os"
	"sync"
	"time"

	"centipede/config"
	"centipede/items"
	"centipede/logs"
)

var log = logs.New()

type (
	Pipeline struct {
		DataChan  chan items.DataRow
		FileChan  chan items.FileRow
		dataLock  sync.RWMutex
		CacheSize int //缓存数量
		//OutPut output.Output
		//ruleTree *spider.RuleTree
	}

	DataCache []Data

	Data map[string]string
)

var appConfig = config.Get()

var fileOutPath = appConfig.FilePath

func (pipeline *Pipeline) AddData(data []items.Data, collection string) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()

	t := time.Now()

	times := items.Data{
		Field: "time",
		Value: t.Format("2006-01-02 15:04:05"),
	}

	data = append(data, times)

	pipeline.DataChan <- items.DataRow{
		collection,
		data,
	}
}

func New() *Pipeline {
	//out := &output.OutputConsole{}

	return &Pipeline{
		DataChan:  make(chan items.DataRow, 100),
		FileChan:  make(chan items.FileRow, 100),
		CacheSize: 4,
		//OutPut:out,
	}
}

func (pipeline *Pipeline) Run(crawler items.CrawlerEr) {

	log.Debugln(crawler.Option().Name, "Pipeline Run")

	defer func() {
		log.Debugln(crawler.Option().Name, " Pipeline", "End")
	}()

	//dataCache := make(items.DataCache,0,pipeline.CacheSize)

	//go func() {

	//TODO
	//Pipeline chanel 没有销毁
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("err")
			}

			log.Debugln(crawler.Option().Name, " Pipeline", "End")

		}()

		for data := range pipeline.DataChan {

			//dataCache = append(dataCache,data)
			//
			//if len(dataCache) < pipeline.CacheSize{
			//	continue
			//}
			//
			//pipeline.OutPut.OutPut(dataCache)
			//
			//dataCache = dataCache[:0]

			//log.Debugln("pipeline", data)

			crawler.Pipeline(data)
		}

	}()

	go func() {
		log.Debug("pipeline.FileChan:", len(pipeline.FileChan))

		var wait sync.WaitGroup

		for file := range pipeline.FileChan {

			wait.Add(1)
			go func() {

				defer func() {
					file.Response.Body.Close()
					wait.Done()
				}()

				d, err := os.Stat(fileOutPath + file.Path)

				if err != nil || !d.IsDir() {
					if err := os.MkdirAll(fileOutPath+file.Path, 0777); err != nil {
						//logs.Log.Error(
						//	" *     Fail  [文件下载：%v | KEYIN：%v | 批次：%v]   %v [ERROR]  %v\n",
						//	self.Spider.GetName(), self.Spider.GetKeyin(), atomic.LoadUint64(&self.fileBatch), fileName, err,
						//)

						log.Error("[创建目录]", err)
						return
					}
				}

				// 文件不存在就以0777的权限创建文件，如果存在就在写入之前清空内容
				f, err := os.OpenFile(fileOutPath+file.Path+"/"+file.FileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
				if err != nil {
					//logs.Log.Error(
					//	" *     Fail  [文件下载：%v | KEYIN：%v | 批次：%v]   %v [ERROR]  %v\n",
					//	self.Spider.GetName(), self.Spider.GetKeyin(), atomic.LoadUint64(&self.fileBatch), fileName, err,
					//)
					log.Error("[文件下载]", err)
					return
				}

				//log.Debugf("存储时: %p",&file.Body,file.Body.ContentLength)
				size, err := io.Copy(f, file.Response.Body)

				log.Debug(file.FileName, "Size:", size)

				if err != nil {
					//logs.Log.Error(
					//	" *     Fail  [文件下载：%v | KEYIN：%v | 批次：%v]   %v (%s) [ERROR]  %v\n",
					//	self.Spider.GetName(), self.Spider.GetKeyin(), atomic.LoadUint64(&self.fileBatch), fileName, bytesSize.Format(uint64(size)), err,
					//)

					log.Error(err)
					return
				}
			}()

			wait.Wait()
		}

	}()

	//<-pipeline.DataChan
	//<-pipeline.FileChan
	//}()

}
