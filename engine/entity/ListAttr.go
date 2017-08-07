package entity

import (
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/typeconv"
)

// List attribute for attribute lists
type ListAttr struct {
	owner  *Entity
	parent interface{}
	pkey   interface{} // key of this item in parent
	path   []interface{}
	flag   attrFlag
	items  []interface{}
}

// Size of list attribute
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

// Set item value
func (a *ListAttr) Set(index int, val interface{}) {
	a.items[index] = val
	if sa, ok := val.(*MapAttr); ok {
		// val is ListAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in index %d", index)
		}

		sa.parent = a
		sa.owner = a.owner
		sa.pkey = index
		sa.flag = a.flag

		a.sendListAttrChangeToClients(index, sa.ToMap())
	} else if sa, ok := val.(*ListAttr); ok {
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in index %d", index)
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

// Get item value
func (a *ListAttr) Get(index int) interface{} {
	val := a.items[index]
	return val
}

// Get item value as int
func (a *ListAttr) GetInt(index int) int {
	val := a.Get(index)
	return int(typeconv.Int(val))
}

// Get item value as int64
func (a *ListAttr) GetInt64(index int) int64 {
	val := a.Get(index)
	return typeconv.Int(val)
}

// Get item value as uint64
func (a *ListAttr) GetUint64(index int) uint64 {
	val := a.Get(index)
	return uint64(typeconv.Int(val))
}

// Get item value as string
func (a *ListAttr) GetStr(index int) string {
	val := a.Get(index)
	return val.(string)
}

// Get item value as float64
func (a *ListAttr) GetFloat(index int) float64 {
	val := a.Get(index)
	return val.(float64)
}

// Get item value as bool
func (a *ListAttr) GetBool(index int) bool {
	val := a.Get(index)
	return val.(bool)
}

// Get item value as List Attribute
func (a *ListAttr) GetListAttr(index int) *ListAttr {
	val := a.Get(index)
	return val.(*ListAttr)
}

// Pop the last item from the end
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

// Pop the last item as ListAttr
func (a *ListAttr) PopListAttr() *ListAttr {
	val := a.Pop()
	return val.(*ListAttr)
}

// Append item to the end
func (a *ListAttr) Append(val interface{}) {
	a.items = append(a.items, val)
	index := len(a.items) - 1

	if sa, ok := val.(*MapAttr); ok {
		// val is ListAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in append")
		}

		sa.parent = a
		sa.owner = a.owner
		sa.pkey = index
		sa.flag = a.flag

		a.sendListAttrAppendToClients(sa.ToMap())
	} else if sa, ok := val.(*ListAttr); ok {
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in append")
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

// Convert ListAttr to slice
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

// Assign slice to ListAttr
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

// Create a new ListAttr
func NewListAttr() *ListAttr {
	return &ListAttr{
		items: []interface{}{},
	}
}
