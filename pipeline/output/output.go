package output

import "centipede/items"

type Output interface {
	OutPut(cache items.DataCache)
}
