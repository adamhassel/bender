package helpers

import (
	"math/rand"
	"time"
)

type StringSlice []string

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (s StringSlice) Random() string {
	if len(s) == 0 {
		return ""
	}
	return s[rand.Intn(len(s))]
}
