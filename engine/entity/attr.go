package entity

type attrFlag int

const (
	afClient attrFlag = 1 << iota
	afAllClient
)

func getPathFromOwner(a interface{}, path []interface{}) []interface{} {
	for {
		if ma, ok := a.(*MapAttr); ok {
			if ma.parent != nil {
				path = append(path, ma.pkey)
				a = ma.parent
			} else {
				break
			}
		} else {
			la := a.(*ListAttr)
			if la.parent != nil {
				path = append(path, la.pkey)
				a = la.parent
			} else {
				break
			}
		}
	}

	return path
}
