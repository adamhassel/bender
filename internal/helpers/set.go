package helpers

type StringSet map[string]struct{}

func NewStringSet(values ...string) StringSet {
	v := make(StringSet)
	v.Add(values...)
	return v
}

func (s StringSet) Add(values ...string) {
	for _, v := range values {
		s[v] = struct{}{}
	}
}

func (s StringSet) Exists(value string) bool {
	_, ok := s[value]
	return ok
}

// Slice returns the values of the set as a slice.
func (s StringSet) Slice() []string {
	if s == nil {
		return nil
	}
	var r = make([]string, len(s))
	i := 0
	for k := range s {
		r[i] = k
		i++
	}
	return r
}

// Remove removes a value from the set.
func (s StringSet) Delete(val string) {
	if s == nil {
		return
	}
	delete(s, val)
}
