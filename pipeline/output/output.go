package output

import "github.com/LSvKing/centipede/items"

type Output interface {
	OutPut(cache items.DataCache)
}
