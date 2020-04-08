package factoids

import (
	"math/rand"
	"time"
)

type factoid struct {
	Value     string    `yaml:"value" json:"value"`
	Origin    string    `yaml:"origin,omitempty" json:"origin,omitempty"`
	SplitWord string    `yaml:"splitword,omitempty" json:"splitword,omitempty"`
	Created   time.Time `yaml:"created,omitempty" json:"created,omitempty"`
}

type FactoidSet map[string]factoid

func NewFactoidSet(values ...factoid) FactoidSet {
	v := make(FactoidSet)
	v.Add(values...)
	return v
}

func (f FactoidSet) Add(values ...factoid) {
	for _, v := range values {
		f[v.Value] = v
	}
}

func (f FactoidSet) Exists(value factoid) bool {
	_, ok := f[value.Value]
	return ok
}

// Slice returns the values of the set as a slice.
func (f FactoidSet) Slice() []factoid {
	if f == nil {
		return nil
	}
	var r = make([]factoid, len(f))
	i := 0
	for _, v := range f {
		r[i] = v
		i++
	}
	return r
}

// Slice returns the values of the set as a slice.
func (f FactoidSet) StringSlice() []string {
	if f == nil {
		return nil
	}
	var r = make([]string, len(f))
	i := 0
	for _, v := range f {
		r[i] = v.Value
		i++
	}
	return r
}

// Remove removes a value from the set.
func (f FactoidSet) Delete(val string) {
	if f == nil {
		return
	}
	delete(f, val)
}

func (f FactoidSet) Random() factoid {
	if len(f) == 0 {
		return factoid{}
	}
	s := make([]string, len(f))
	i := 0
	for k := range f {
		s[i] = k
		i++
	}
	return f[s[rand.Intn(len(s))]]
}
