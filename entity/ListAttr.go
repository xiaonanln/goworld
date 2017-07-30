package entity

import (
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/typeconv"
)

type ListAttr struct {
	owner  *Entity
	parent interface{}
	pkey   interface{} // key of this item in parent
	path   []interface{}
	flag   attrFlag
	items  []interface{}
}

func (a *ListAttr) Size() int {
	return len(a.items)
}

func (a *ListAttr) clearOwner() {
	a.owner = nil
	a.parent = nil
	a.pkey = nil
	a.path = nil
	a.flag = 0
}

func (a *ListAttr) Set(index int, val interface{}) {
	a.items[index] = val
	if sa, ok := val.(*MapAttr); ok {
		// val is ListAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in index %s", index)
		}

		sa.parent = a
		sa.owner = a.owner
		sa.pkey = index
		sa.flag = a.flag

		a.sendListAttrChangeToClients(index, sa.ToMap())
	} else if sa, ok := val.(*ListAttr); ok {
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in index %s", index)
		}

		sa.parent = a
		sa.owner = a.owner
		sa.pkey = index
		sa.flag = a.flag

		a.sendListAttrChangeToClients(index, sa.ToList())
	} else {
		a.sendListAttrChangeToClients(index, val)
	}
}

func (a *ListAttr) sendListAttrChangeToClients(index int, val interface{}) {
	owner := a.owner
	if owner != nil {
		// send the change to owner's client
		owner.sendListAttrChangeToClients(a, index, val)
	}
}

func (a *ListAttr) sendListAttrPopToClients() {
	if owner := a.owner; owner != nil {
		owner.sendListAttrPopToClients(a)
	}
}

func (a *ListAttr) sendListAttrAppendToClients(val interface{}) {
	if owner := a.owner; owner != nil {
		owner.sendListAttrAppendToClients(a, val)
	}
}

func (a *ListAttr) getPathFromOwner() []interface{} {
	if a.path == nil {
		a.path = a._getPathFromOwner()
	}
	return a.path
}

func (a *ListAttr) _getPathFromOwner() []interface{} {
	path := make([]interface{}, 0, 4)
	if a.parent != nil {
		path = append(path, a.pkey)
		return getPathFromOwner(a.parent, path)
	} else {
		return path
	}
}

func (a *ListAttr) Get(index int) interface{} {
	val := a.items[index]
	return val
}

func (a *ListAttr) GetInt(index int) int {
	val := a.Get(index)
	return int(typeconv.Int(val))
}

func (a *ListAttr) GetInt64(index int) int64 {
	val := a.Get(index)
	return typeconv.Int(val)
}

func (a *ListAttr) GetUint64(index int) uint64 {
	val := a.Get(index)
	return uint64(typeconv.Int(val))
}

func (a *ListAttr) GetStr(index int) string {
	val := a.Get(index)
	return val.(string)
}

func (a *ListAttr) GetFloat(index int) float64 {
	val := a.Get(index)
	return val.(float64)
}

func (a *ListAttr) GetBool(index int) bool {
	val := a.Get(index)
	return val.(bool)
}

func (a *ListAttr) GetListAttr(index int) *ListAttr {
	val := a.Get(index)
	return val.(*ListAttr)
}

// Delete a key in attrs
func (a *ListAttr) Pop() interface{} {
	size := len(a.items)
	val := a.items[size-1]
	a.items = a.items[:size-1]

	if sa, ok := val.(*MapAttr); ok {
		sa.clearOwner()
	} else if sa, ok := val.(*ListAttr); ok {
		sa.clearOwner()
	}

	a.sendListAttrPopToClients()
	return val
}

func (a *ListAttr) PopListAttr() *ListAttr {
	val := a.Pop()
	return val.(*ListAttr)
}

func (a *ListAttr) Append(val interface{}) {
	a.items = append(a.items, val)
	index := len(a.items) - 1

	if sa, ok := val.(*MapAttr); ok {
		// val is ListAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in append", index)
		}

		sa.parent = a
		sa.owner = a.owner
		sa.pkey = index
		sa.flag = a.flag

		a.sendListAttrAppendToClients(sa.ToMap())
	} else if sa, ok := val.(*ListAttr); ok {
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in append", index)
		}

		sa.parent = a
		sa.owner = a.owner
		sa.pkey = index
		sa.flag = a.flag

		a.sendListAttrAppendToClients(sa.ToList())
	} else {
		a.sendListAttrAppendToClients(val)
	}
}

func (a *ListAttr) ToList() []interface{} {
	l := make([]interface{}, len(a.items))

	for i, v := range a.items {
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

func (a *ListAttr) AssignList(l []interface{}) {
	for _, v := range l {
		if iv, ok := v.(map[string]interface{}); ok {
			ia := NewMapAttr()
			ia.AssignMap(iv)
			a.Append(ia)
		} else if iv, ok := v.([]interface{}); ok {
			ia := NewListAttr()
			ia.AssignList(iv)
			a.Append(ia)
		} else {
			a.Append(v)
		}
	}
}

//func (a *ListAttr) AssignList(l []interface{}) *ListAttr {
//
//	for idx, v := l {
//		innerMap, ok := v.(map[string]interface{})
//		if ok {
//			innerListAttr := NewListAttr()
//			innerListAttr.AssignMap(innerMap)
//			a.Set(idx, innerListAttr)
//		} else {
//			a.Set(idx, v)
//		}
//	}
//	return a
//}

func NewListAttr() *ListAttr {
	return &ListAttr{
		items: []interface{}{},
	}
}
