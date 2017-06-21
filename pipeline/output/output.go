package output

import "douban_spider/items"

type Output interface {
	OutPut(cache items.DataCache)
}
