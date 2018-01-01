package common

import (
	"math/rand"
	"strconv"
)

func GetProxy() string {
	return strconv.Itoa(rand.Int())
}
