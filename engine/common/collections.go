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
