package common

// StringSet is a set of strings
type StringSet map[string]struct{}

// Contains checks if Stringset contains the string
func (ss StringSet) Contains(elem string) bool {
	_, ok := ss[elem]
	return ok
}

// Add adds the string to StringSet
func (ss StringSet) Add(elem string) {
	ss[elem] = struct{}{}
}

// Remove removes the string from StringList
func (ss StringSet) Remove(elem string) {
	delete(ss, elem)
}

// ToList convert StringSet to string slice
func (ss StringSet) ToList() []string {
	keys := make([]string, 0, len(ss))
	for s := range ss {
		keys = append(keys, s)
	}
	return keys
}

// StringList is a list of string (slice)
type StringList []string

// Remove removes the string from StringList
func (sl *StringList) Remove(elem string) {
	widx := 0
	cpsl := *sl
	for idx, _elem := range cpsl {
		if _elem == elem {
			// ignore this elem by doing nothing
		} else {
			if idx != widx {
				cpsl[widx] = _elem
			}
			widx += 1
		}
	}

	*sl = cpsl[:widx]
}

// Append add the string to the end of StringList
func (sl *StringList) Append(elem string) {
	*sl = append(*sl, elem)
}

// Find get the index of string in StringList, returns -1 if not found
func (sl *StringList) Find(s string) int {
	for idx, elem := range *sl {
		if elem == s {
			return idx
		}
	}
	return -1
}

// IntSet is a set of int
type IntSet map[int]struct{}

// Contains checks if Stringset contains the string
func (is IntSet) Contains(elem int) bool {
	_, ok := is[elem]
	return ok
}

// Add adds the string to IntSet
func (is IntSet) Add(elem int) {
	is[elem] = struct{}{}
}

// Remove removes the string from IntSet
func (is IntSet) Remove(elem int) {
	delete(is, elem)
}

// ToList convert IntSet to int slice
func (is IntSet) ToList() []int {
	keys := make([]int, 0, len(is))
	for s := range is {
		keys = append(keys, s)
	}
	return keys
}

// Uint16Set is a set of int
type Uint16Set map[uint16]struct{}

// Contains checks if Stringset contains the string
func (is Uint16Set) Contains(elem uint16) bool {
	_, ok := is[elem]
	return ok
}

// Add adds the string to Uint16Set
func (is Uint16Set) Add(elem uint16) {
	is[elem] = struct{}{}
}

// Remove removes the string from Uint16Set
func (is Uint16Set) Remove(elem uint16) {
	delete(is, elem)
}

// ToList convert Uint16Set to int slice
func (is Uint16Set) ToList() []uint16 {
	keys := make([]uint16, 0, len(is))
	for s := range is {
		keys = append(keys, s)
	}
	return keys
}
