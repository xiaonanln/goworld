package entity

import (
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/typeconv"
)

// MapAttr is a map attribute containing muiltiple attributes indexed by string keys
type MapAttr struct {
	owner  *Entity
	parent interface{}
	pkey   interface{} // key of this item in parent
	path   []interface{}
	flag   attrFlag
	attrs  map[string]interface{}
}

// Size returns the size of MapAttr
func (a *MapAttr) Size() int {
	return len(a.attrs)
}

// HasKey returns if the key exists in MapAttr
func (a *MapAttr) HasKey(key string) bool {
	_, ok := a.attrs[key]
	return ok
}

// Set sets the key-attribute pair in MapAttr
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

// SetDefault sets the key-attribute pair in MapAttr if key not exists
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
	if a.parent == nil {
		return nil
	}

	path := make([]interface{}, 0, 4)
	path = append(path, a.pkey)
	return getPathFromOwner(a.parent, path)
}

// Get returns the attribute of specified key in MapAttr
func (a *MapAttr) Get(key string) interface{} {
	val, ok := a.attrs[key]
	if !ok {
		gwlog.Panicf("key not exists: %s", key)
	}
	return val
}

// GetInt returns the attribute of specified key in MapAttr as int
func (a *MapAttr) GetInt(key string) int {
	val := a.Get(key)
	return int(typeconv.Int(val))
}

// GetInt64 returns the attribute of specified key in MapAttr as int64
func (a *MapAttr) GetInt64(key string) int64 {
	val := a.Get(key)
	return typeconv.Int(val)
}

// GetUint64 returns the attribute of specified key in MapAttr as uint64
func (a *MapAttr) GetUint64(key string) uint64 {
	val := a.Get(key)
	return uint64(typeconv.Int(val))
}

// GetStr returns the attribute of specified key in MapAttr as string
func (a *MapAttr) GetStr(key string) string {
	val := a.Get(key)
	return val.(string)
}

// GetFloat returns the attribute of specified key in MapAttr as float64
func (a *MapAttr) GetFloat(key string) float64 {
	val := a.Get(key)
	return val.(float64)
}

// GetBool returns the attribute of specified key in MapAttr as bool
func (a *MapAttr) GetBool(key string) bool {
	val := a.Get(key)
	return val.(bool)
}

// GetMapAttr returns the attribute of specified key in MapAttr as MapAttr
func (a *MapAttr) GetMapAttr(key string) *MapAttr {
	val := a.Get(key)
	return val.(*MapAttr)
}

// GetListAttr returns the attribute of specified key in MapAttr as ListAttr
func (a *MapAttr) GetListAttr(key string) *ListAttr {
	val := a.Get(key)
	return val.(*ListAttr)
}

// Pop deletes a key in MapAttr and returns the attribute
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

// Del deletes a key in MapAttr
func (a *MapAttr) Del(key string) {
	a.Pop(key)
}

// PopMapAttr deletes a key in MapAttr and returns the attribute as MapAttr
func (a *MapAttr) PopMapAttr(key string) *MapAttr {
	val := a.Pop(key)
	return val.(*MapAttr)
}

//// Clear removes all key-values from the MapAttr
//func (a *MapAttr) Clear() {
//
//	val, ok := a.attrs[key]
//	if !ok {
//		gwlog.Panicf("key not exists: %s", key)
//	}
//
//	delete(a.attrs, key)
//	if sa, ok := val.(*MapAttr); ok {
//		sa.clearOwner()
//	} else if sa, ok := val.(*ListAttr); ok {
//		sa.clearOwner()
//	}
//
//	a.sendAttrDelToClients(key)
//	return val
//}

//// GetKeys returns all keys of MapAttr as slice of strings
//func (a *MapAttr) GetKeys() []string {
//	size := len(a.attrs)
//	keys := make([]string, 0, size)
//	for k := range a.attrs {
//		keys = append(keys, k)
//	}
//	return keys
//}
//
//func (a *MapAttr) GetValues() []interface{} {
//	size := len(a.attrs)
//	vals := make([]interface{}, 0, size)
//	for _, v := range a.attrs {
//		vals = append(vals, v)
//	}
//	return vals
//}

// ToMap converts MapAttr to native map, recursively
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

// ToMapWithFilter converts filtered fields of MapAttr to to native map, recursively
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

// AssignMap assigns native map to MapAttr recursively
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

// AssignMapWithFilter assigns filtered fields of native map to MapAttr recursively
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

// NewMapAttr creates a new MapAttr
func NewMapAttr() *MapAttr {
	return &MapAttr{
		attrs: make(map[string]interface{}),
	}
}
