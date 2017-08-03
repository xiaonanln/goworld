package goworld_ds

type StringSet map[string]bool

func (ss StringSet) Contains(elem string) bool {
	return ss[elem]
}

func (ss StringSet) Add(elem string) {
	ss[elem] = true
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

func init() {
}
