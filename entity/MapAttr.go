package entity

import (
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/typeconv"
)

type MapAttr struct {
	owner  *Entity
	parent interface{}
	pkey   interface{} // key of this item in parent
	path   []interface{}
	flag   attrFlag
	attrs  map[string]interface{}
}

func (a *MapAttr) Size() int {
	return len(a.attrs)
}

func (a *MapAttr) HasKey(key string) bool {
	_, ok := a.attrs[key]
	return ok
}

func (a *MapAttr) Set(key string, val interface{}) {
	a.attrs[key] = val
	if sa, ok := val.(*MapAttr); ok {
		// val is MapAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in key %s", key)
		}

		sa.parent = a
		sa.owner = a.owner
		sa.pkey = key
		if a == a.owner.Attrs { // this is the root
			sa.flag = a.owner.getAttrFlag(key)
		} else {
			sa.flag = a.flag
		}

		a.sendAttrChangeToClients(key, sa.ToMap())
	} else if sa, ok := val.(*ListAttr); ok {
		// val is ListATtr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("ListAttr reused in key %s", key)
		}

		sa.parent = a
		sa.owner = a.owner
		sa.pkey = key
		if a == a.owner.Attrs { // this is the root
			sa.flag = a.owner.getAttrFlag(key)
		} else {
			sa.flag = a.flag
		}

		a.sendAttrChangeToClients(key, sa.ToList())
	} else {
		a.sendAttrChangeToClients(key, val)
	}
}
func (a *MapAttr) SetDefault(key string, val interface{}) {
	if _, ok := a.attrs[key]; !ok {
		a.Set(key, val)
	}
}

func (a *MapAttr) sendAttrChangeToClients(key string, val interface{}) {
	if owner := a.owner; owner != nil {
		// send the change to owner's client
		owner.sendMapAttrChangeToClients(a, key, val)
	}
}

func (a *MapAttr) sendAttrDelToClients(key string) {
	if owner := a.owner; owner != nil {
		owner.sendMapAttrDelToClients(a, key)
	}
}

func (a *MapAttr) getPathFromOwner() []interface{} {
	if a.path == nil {
		a.path = a._getPathFromOwner()
	}
	return a.path
}

func (a *MapAttr) _getPathFromOwner() []interface{} {
	path := make([]interface{}, 0, 4)
	if a.parent != nil {
		path = append(path, a.pkey)
		return getPathFromOwner(a.parent, path)
	} else {
		return path
	}
}

func (a *MapAttr) Get(key string) interface{} {
	val, ok := a.attrs[key]
	if !ok {
		gwlog.Panicf("key not exists: %s", key)
	}
	return val
}

func (a *MapAttr) GetInt(key string) int {
	val := a.Get(key)
	return int(typeconv.Int(val))
}

func (a *MapAttr) GetInt64(key string) int64 {
	val := a.Get(key)
	return typeconv.Int(val)
}

func (a *MapAttr) GetUint64(key string) uint64 {
	val := a.Get(key)
	return uint64(typeconv.Int(val))
}

func (a *MapAttr) GetStr(key string) string {
	val := a.Get(key)
	return val.(string)
}

func (a *MapAttr) GetFloat(key string) float64 {
	val := a.Get(key)
	return val.(float64)
}

func (a *MapAttr) GetBool(key string) bool {
	val := a.Get(key)
	return val.(bool)
}

func (a *MapAttr) GetMapAttr(key string) *MapAttr {
	val := a.Get(key)
	return val.(*MapAttr)
}

func (a *MapAttr) GetListAttr(key string) *ListAttr {
	val := a.Get(key)
	return val.(*ListAttr)
}

// Delete a key in attrs
func (a *MapAttr) Pop(key string) interface{} {
	val, ok := a.attrs[key]
	if !ok {
		gwlog.Panicf("key not exists: %s", key)
	}

	delete(a.attrs, key)
	if sa, ok := val.(*MapAttr); ok {
		sa.clearOwner()
	} else if sa, ok := val.(*ListAttr); ok {
		sa.clearOwner()
	}

	a.sendAttrDelToClients(key)
	return val
}

func (a *MapAttr) Del(key string) {
	a.Pop(key)
}

func (a *MapAttr) PopMapAttr(key string) *MapAttr {
	val := a.Pop(key)
	return val.(*MapAttr)
}

func (a *MapAttr) GetKeys() []string {
	size := len(a.attrs)
	keys := make([]string, 0, size)
	for k, _ := range a.attrs {
		keys = append(keys, k)
	}
	return keys
}

func (a *MapAttr) GetValues() []interface{} {
	size := len(a.attrs)
	vals := make([]interface{}, 0, size)
	for _, v := range a.attrs {
		vals = append(vals, v)
	}
	return vals
}

func (a *MapAttr) ToMap() map[string]interface{} {
	doc := map[string]interface{}{}
	for k, v := range a.attrs {
		if a, ok := v.(*MapAttr); ok {
			doc[k] = a.ToMap()
		} else if a, ok := v.(*ListAttr); ok {
			doc[k] = a.ToList()
		} else {
			doc[k] = v
		}
	}
	return doc
}

func (a *MapAttr) ToMapWithFilter(filter func(string) bool) map[string]interface{} {
	doc := map[string]interface{}{}
	for k, v := range a.attrs {
		if !filter(k) {
			continue
		}

		if a, ok := v.(*MapAttr); ok {
			doc[k] = a.ToMap()
		} else if a, ok := v.(*ListAttr); ok {
			doc[k] = a.ToList()
		} else {
			doc[k] = v
		}
	}
	return doc
}

func (a *MapAttr) AssignMap(doc map[string]interface{}) {
	for k, v := range doc {
		if iv, ok := v.(map[string]interface{}); ok {
			ia := NewMapAttr()
			ia.AssignMap(iv)
			a.Set(k, ia)
		} else if iv, ok := v.([]interface{}); ok {
			ia := NewListAttr()
			ia.AssignList(iv)
			a.Set(k, ia)
		} else {
			a.Set(k, v)
		}
	}
}

func (a *MapAttr) AssignMapWithFilter(doc map[string]interface{}, filter func(string) bool) {
	for k, v := range doc {
		if !filter(k) {
			continue
		}

		if iv, ok := v.(map[string]interface{}); ok {
			ia := NewMapAttr()
			ia.AssignMap(iv)
			a.Set(k, ia)
		} else if iv, ok := v.([]interface{}); ok {
			ia := NewListAttr()
			ia.AssignList(iv)
			a.Set(k, ia)
		} else {
			a.Set(k, v)
		}
	}
}

func (a *MapAttr) clearOwner() {
	a.owner = nil
	a.parent = nil
	a.pkey = nil
	a.path = nil
	a.flag = 0
}

func NewMapAttr() *MapAttr {
	return &MapAttr{
		attrs: make(map[string]interface{}),
	}
}
