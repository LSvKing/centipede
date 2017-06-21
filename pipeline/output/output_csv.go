package output

import (
	"douban_spider/items"
	"douban_spider/logs"
	"encoding/csv"
	"fmt"
	"os"
)

type (
	OutPutCSV struct {
	}

	handle map[string]struct {
		file   *os.File
		writer *csv.Writer
	}
)

var log = logs.New()

func (this *OutPutCSV) OutPut(dataCache items.DataCache) {
	fmt.Println("csv")

	handle := make(handle)

	defer func() {
		if len(handle) > 0 {
			for _, handler := range handle {
				handler.writer.Flush()
				handler.file.Close()
			}
		}
	}()

	for _, value := range dataCache {
		var data [][]string
		var dataItem []string

		if _, ok := handle[value.Collection]; !ok {
			file, err := os.OpenFile(value.Collection+".csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				log.Fatal(err)
			}
			file.WriteString("\xEF\xBB\xBF")

			handle[value.Collection] = struct {
				file   *os.File
				writer *csv.Writer
			}{file: file, writer: csv.NewWriter(file)}
		}

		for _, v := range value.Data {
			dataItem = append(dataItem, v.Value.(string))
		}

		data = append(data, dataItem)

		handle[value.Collection].writer.WriteAll(data)
	}
}
