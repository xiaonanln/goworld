package common

type StringSet map[string]struct{}

func (ss StringSet) Contains(elem string) bool {
	_, ok := ss[elem]
	return ok
}

func (ss StringSet) Add(elem string) {
	ss[elem] = struct{}{}
}

func (ss StringSet) Remove(elem string) {
	delete(ss, elem)
}

type StringList []string

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

func (sl *StringList) Append(elem string) {
	*sl = append(*sl, elem)
}

func (sl *StringList) Find(s string) int {
	for idx, elem := range *sl {
		if elem == s {
			return idx
		}
	}
	return -1
}
