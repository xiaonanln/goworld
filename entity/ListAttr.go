package entity

import (
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/typeconv"
)

type ListAttr struct {
	owner  *Entity
	parent interface{}
	pkey   interface{} // key of this item in parent
	items  []interface{}
}

func (la *ListAttr) Size() int {
	return len(la.items)
}

func (la *ListAttr) clearOwner() {
	la.owner = nil
	la.parent = nil
	la.pkey = nil
}

func (la *ListAttr) Set(index int, val interface{}) {
	la.items[index] = val
	if sa, ok := val.(*MapAttr); ok {
		// val is ListAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in index %s", index)
		}

		sa.parent = la
		sa.owner = la.owner
		sa.pkey = index

		la.sendListAttrChangeToClients(index, sa.ToMap())
	} else if sa, ok := val.(*ListAttr); ok {
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in index %s", index)
		}

		sa.parent = la
		sa.owner = la.owner
		sa.pkey = index

		la.sendListAttrChangeToClients(index, sa.ToList())
	} else {
		la.sendListAttrChangeToClients(index, val)
	}
}

func (la *ListAttr) sendListAttrChangeToClients(index int, val interface{}) {
	owner := la.owner
	if owner != nil {
		// send the change to owner's client
		owner.sendListAttrChangeToClients(la, index, val)
	}
}

func (la *ListAttr) sendListAttrPopToClients() {
	if owner := la.owner; owner != nil {
		owner.sendListAttrPopToClients(la)
	}
}

func (la *ListAttr) sendListAttrAppendToClients(val interface{}) {
	if owner := la.owner; owner != nil {
		owner.sendListAttrAppendToClients(la, val)
	}
}

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

func (la *ListAttr) getPathFromOwner() []interface{} {
	path := make([]interface{}, 0, 4)
	if la.parent != nil {
		path = append(path, la.pkey)
		return getPathFromOwner(la.parent, path)
	} else {
		return path
	}
}

func (la *ListAttr) Get(index int) interface{} {
	val := la.items[index]
	return val
}

func (la *ListAttr) GetInt(index int) int {
	val := la.Get(index)
	return int(typeconv.Int(val))
}

func (la *ListAttr) GetInt64(index int) int64 {
	val := la.Get(index)
	return typeconv.Int(val)
}

func (la *ListAttr) GetUint64(index int) uint64 {
	val := la.Get(index)
	return uint64(typeconv.Int(val))
}

func (la *ListAttr) GetStr(index int) string {
	val := la.Get(index)
	return val.(string)
}

func (la *ListAttr) GetFloat(index int) float64 {
	val := la.Get(index)
	return val.(float64)
}

func (la *ListAttr) GetBool(index int) bool {
	val := la.Get(index)
	return val.(bool)
}

func (la *ListAttr) GetListAttr(index int) *ListAttr {
	val := la.Get(index)
	return val.(*ListAttr)
}

// Delete a key in attrs
func (la *ListAttr) Pop() interface{} {
	size := len(la.items)
	val := la.items[size-1]
	la.items = la.items[:size-1]

	if sa, ok := val.(*MapAttr); ok {
		sa.clearOwner()
	} else if sa, ok := val.(*ListAttr); ok {
		sa.clearOwner()
	}

	la.sendListAttrPopToClients()
	return val
}

func (la *ListAttr) PopListAttr() *ListAttr {
	val := la.Pop()
	return val.(*ListAttr)
}

func (la *ListAttr) Append(val interface{}) {
	la.items = append(la.items, val)
	index := len(la.items) - 1

	if sa, ok := val.(*MapAttr); ok {
		// val is ListAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in append", index)
		}

		sa.parent = la
		sa.owner = la.owner
		sa.pkey = index

		la.sendListAttrAppendToClients(sa.ToMap())
	} else if sa, ok := val.(*ListAttr); ok {
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in append", index)
		}

		sa.parent = la
		sa.owner = la.owner
		sa.pkey = index

		la.sendListAttrAppendToClients(sa.ToList())
	} else {
		la.sendListAttrAppendToClients(val)
	}
}

func (la *ListAttr) ToList() []interface{} {
	l := make([]interface{}, len(la.items))

	for i, v := range la.items {
		if ma, ok := v.(*MapAttr); ok {
			l[i] = ma.ToMap()
		} else if la, ok := v.(*ListAttr); ok {
			l[i] = la.ToList()
		} else {
			l[i] = v
		}
	}
	return l
}

func (la *ListAttr) AssignList(l []interface{}) {
	for _, v := range l {
		if iv, ok := v.(map[string]interface{}); ok {
			ia := NewMapAttr()
			ia.AssignMap(iv)
			la.Append(ia)
		} else if iv, ok := v.([]interface{}); ok {
			ia := NewListAttr()
			ia.AssignList(iv)
			la.Append(ia)
		} else {
			la.Append(v)
		}
	}
}

//func (la *ListAttr) AssignList(l []interface{}) *ListAttr {
//
//	for idx, v := l {
//		innerMap, ok := v.(map[string]interface{})
//		if ok {
//			innerListAttr := NewListAttr()
//			innerListAttr.AssignMap(innerMap)
//			la.Set(idx, innerListAttr)
//		} else {
//			la.Set(idx, v)
//		}
//	}
//	return la
//}

func NewListAttr() *ListAttr {
	return &ListAttr{
		items: []interface{}{},
	}
}
