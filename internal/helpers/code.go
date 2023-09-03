package helpers

import (
	"fmt"
	"math/rand"
	"time"
)

func GenCode() string {
	rand.Seed(time.Now().UnixNano())
	var availableNumbers [3]int
	for i := 0; i < 3; i++ {
		availableNumbers[i] = rand.Intn(9)
	}
	var code string
	for i := 0; i < 6; i++ {
		randNum := availableNumbers[rand.Intn(len(availableNumbers))]

		code = fmt.Sprintf("%s%d", code, randNum)
	}
	return code
}
