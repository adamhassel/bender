package helpers

import (
	"math/rand"
)

type Slice[T any] []T

func (s Slice[T]) Random() (rv T) {
	if len(s) == 0 {
		return rv
	}
	return s[rand.Intn(len(s))]
}
