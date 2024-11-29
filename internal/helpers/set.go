package helpers

import "math/rand"

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](values ...T) Set[T] {
	v := make(Set[T])
	v.Add(values...)
	return v
}

func (s Set[T]) Add(values ...T) {
	for _, v := range values {
		s[v] = struct{}{}
	}
}

func (s Set[T]) Exists(value T) bool {
	_, ok := s[value]
	return ok
}

// Random returns a random member of s
func (s Set[T]) Random() (rv T) {
	n := rand.Intn(len(s))
	i := 0
	for k := range s {
		if i == n {
			return k
		}
		i++
	}
	return
}

// Slice returns the values of the set as a slice.
func (s Set[T]) Slice() Slice[T] {
	if s == nil {
		return nil
	}
	var r = make(Slice[T], len(s))
	i := 0
	for k := range s {
		r[i] = k
		i++
	}
	return r
}

// Delete removes a value from the set.
func (s Set[T]) Delete(val T) {
	if s == nil {
		return
	}
	delete(s, val)
}
