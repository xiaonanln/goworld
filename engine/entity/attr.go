package entity

import "github.com/xiaonanln/goworld/engine/gwlog"

type attrFlag int

const (
	afClient attrFlag = 1 << iota
	afAllClient
)

func getPathFromOwner(a interface{}, path []interface{}) []interface{} {
forloop:
	for {
		switch ma := a.(type) {
		case *MapAttr:
			if ma.parent != nil {
				path = append(path, ma.pkey)
				a = ma.parent
			} else {
				break forloop
			}
		case *ListAttr:
			if ma.parent != nil {
				path = append(path, ma.pkey)
				a = ma.parent
			} else {
				break forloop
			}
		default:
			gwlog.Panicf("getPathFromOwner: invalid parent type: %T", a)
		}
	}

	return path
}

// uniformAttrType convert v to uniform attr type: int64, float64, bool, string
func uniformAttrType(v interface{}) interface{} {
	switch av := v.(type) {
	case bool:
		return av
	case string:
		return av
	case float64:
		return float64(av)
	case float32:
		return float64(av)
	case int64:
		return av
	case uint64:
		return int64(av)
	case int:
		return int64(av)
	case uint:
		return int64(av)
	case int32:
		return int64(av)
	case uint32:
		return int64(av)
	case int16:
		return int64(av)
	case uint16:
		return int64(av)
	case int8:
		return int64(av)
	case byte:
		return int64(av)

	default:
		gwlog.Panicf("can not uniform attr val %+v type %T", v, v)
		return v
	}

}
