package sensors

import (
	"math/rand"
	"time"
)

var (
	theRand *rand.Rand
)

func init() {
	theRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}
