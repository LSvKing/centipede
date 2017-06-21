package output

import (
	"douban_spider/items"
	"fmt"
)

type OutputConsole struct {
}

func (this *OutputConsole) OutPut(dataCache items.DataCache) {
	log.Debug("Console")

	for _, value := range dataCache {
		fmt.Println("-----------------" + value.Collection + "-------------------")
		for _, v := range value.Data {
			//fmt.Println("Field:" + v.Field + " => " + "Value:" + v.Value )
			fmt.Println(v)
		}
	}
}
