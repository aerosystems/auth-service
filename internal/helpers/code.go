package helpers

import (
	"math/rand"
	"time"
)

func GenCode() int {
	rand.Seed(time.Now().UnixNano())
	min := 100000
	max := 999999
	return rand.Intn(max-min+1) + min
}
