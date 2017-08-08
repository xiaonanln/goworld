package gwds

// StringSet is a data structure for a set of strings
type StringSet map[string]bool

// Contains check if a string is in StringSet
func (ss StringSet) Contains(elem string) bool {
	return ss[elem]
}

// Add adds a string to StringSet
func (ss StringSet) Add(elem string) {
	ss[elem] = true
}

// Remove removes a string from StringSet
func (ss StringSet) Remove(elem string) {
	delete(ss, elem)
}

// StringList is a list of strings
type StringList []string

// Remove removes a string from StringList
func (sl *StringList) Remove(elem string) {
	widx := 0
	cpsl := *sl
	for idx, _elem := range cpsl {
		if _elem == elem {
			// ignore this elem by doing nothing
		} else if idx != widx {
			cpsl[widx] = _elem
			widx += 1
		}
	}

	*sl = cpsl[:widx]
}

// Append appends a string to the end of StringList
func (sl *StringList) Append(elem string) {
	*sl = append(*sl, elem)
}

// Find finds the index of string in StringList, or -1 if not found
func (sl *StringList) Find(s string) int {
	for idx, elem := range *sl {
		if elem == s {
			return idx
		}
	}
	return -1
}
