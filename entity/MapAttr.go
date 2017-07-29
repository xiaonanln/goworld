package entity

import (
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/typeconv"
)

type MapAttr struct {
	owner  *Entity
	parent interface{}
	pkey   interface{} // key of this item in parent
	attrs  map[string]interface{}
}

func (ma *MapAttr) Size() int {
	return len(ma.attrs)
}

func (ma *MapAttr) HasKey(key string) bool {
	_, ok := ma.attrs[key]
	return ok
}

func (ma *MapAttr) Set(key string, val interface{}) {
	ma.attrs[key] = val
	if sa, ok := val.(*MapAttr); ok {
		// val is MapAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in key %s", key)
		}

		sa.parent = ma
		sa.owner = ma.owner
		sa.pkey = key

		ma.sendAttrChangeToClients(key, sa.ToMap())
	} else if sa, ok := val.(*ListAttr); ok {
		// val is ListATtr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("ListAttr reused in key %s", key)
		}

		sa.parent = ma
		sa.owner = ma.owner
		sa.pkey = key

		ma.sendAttrChangeToClients(key, sa.ToList())
	} else {
		ma.sendAttrChangeToClients(key, val)
	}
}
func (ma *MapAttr) SetDefault(key string, val interface{}) {
	if _, ok := ma.attrs[key]; !ok {
		ma.Set(key, val)
	}
}

func (ma *MapAttr) sendAttrChangeToClients(key string, val interface{}) {
	if owner := ma.owner; owner != nil {
		// send the change to owner's client
		owner.sendMapAttrChangeToClients(ma, key, val)
	}
}

func (ma *MapAttr) sendAttrDelToClients(key string) {
	if owner := ma.owner; owner != nil {
		owner.sendMapAttrDelToClients(ma, key)
	}
}

func (ma *MapAttr) getPathFromOwner() []interface{} {
	path := make([]interface{}, 0, 4)
	if ma.parent != nil {
		path = append(path, ma.pkey)
		return getPathFromOwner(ma.parent, path)
	} else {
		return path
	}
}

func (ma *MapAttr) Get(key string) interface{} {
	val, ok := ma.attrs[key]
	if !ok {
		gwlog.Panicf("key not exists: %s", key)
	}
	return val
}

func (ma *MapAttr) GetInt(key string) int {
	val := ma.Get(key)
	return int(typeconv.Int(val))
}

func (ma *MapAttr) GetInt64(key string) int64 {
	val := ma.Get(key)
	return typeconv.Int(val)
}

func (ma *MapAttr) GetUint64(key string) uint64 {
	val := ma.Get(key)
	return uint64(typeconv.Int(val))
}

func (ma *MapAttr) GetStr(key string) string {
	val := ma.Get(key)
	return val.(string)
}

func (ma *MapAttr) GetFloat(key string) float64 {
	val := ma.Get(key)
	return val.(float64)
}

func (ma *MapAttr) GetBool(key string) bool {
	val := ma.Get(key)
	return val.(bool)
}

func (ma *MapAttr) GetMapAttr(key string) *MapAttr {
	val := ma.Get(key)
	return val.(*MapAttr)
}

func (ma *MapAttr) GetListAttr(key string) *ListAttr {
	val := ma.Get(key)
	return val.(*ListAttr)
}

// Delete a key in attrs
func (ma *MapAttr) Pop(key string) interface{} {
	val, ok := ma.attrs[key]
	if !ok {
		gwlog.Panicf("key not exists: %s", key)
	}

	delete(ma.attrs, key)
	if sa, ok := val.(*MapAttr); ok {
		sa.parent = nil // clear parent and owner when attribute is poped
		sa.pkey = ""
		sa.owner = nil
	} else if sa, ok := val.(*ListAttr); ok {
		sa.parent = nil // clear parent and owner when attribute is poped
		sa.pkey = ""
		sa.owner = nil
	}

	ma.sendAttrDelToClients(key)
	return val
}

func (ma *MapAttr) Del(key string) {
	ma.Pop(key)
}

func (ma *MapAttr) PopMapAttr(key string) *MapAttr {
	val := ma.Pop(key)
	return val.(*MapAttr)
}

func (ma *MapAttr) GetKeys() []string {
	size := len(ma.attrs)
	keys := make([]string, 0, size)
	for k, _ := range ma.attrs {
		keys = append(keys, k)
	}
	return keys
}

func (ma *MapAttr) GetValues() []interface{} {
	size := len(ma.attrs)
	vals := make([]interface{}, 0, size)
	for _, v := range ma.attrs {
		vals = append(vals, v)
	}
	return vals
}

func (ma *MapAttr) ToMap() map[string]interface{} {
	doc := map[string]interface{}{}
	for k, v := range ma.attrs {
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

func (ma *MapAttr) ToMapWithFilter(filter func(string) bool) map[string]interface{} {
	doc := map[string]interface{}{}
	for k, v := range ma.attrs {
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

func (ma *MapAttr) AssignMap(doc map[string]interface{}) {
	for k, v := range doc {
		if iv, ok := v.(map[string]interface{}); ok {
			ia := NewMapAttr()
			ia.AssignMap(iv)
			ma.Set(k, ia)
		} else if iv, ok := v.([]interface{}); ok {
			ia := NewListAttr()
			ia.AssignList(iv)
			ma.Set(k, ia)
		} else {
			ma.Set(k, v)
		}
	}
}

func (ma *MapAttr) AssignMapWithFilter(doc map[string]interface{}, filter func(string) bool) {
	for k, v := range doc {
		if !filter(k) {
			continue
		}

		if iv, ok := v.(map[string]interface{}); ok {
			ia := NewMapAttr()
			ia.AssignMap(iv)
			ma.Set(k, ia)
		} else if iv, ok := v.([]interface{}); ok {
			ia := NewListAttr()
			ia.AssignList(iv)
			ma.Set(k, ia)
		} else {
			ma.Set(k, v)
		}
	}
}

func (ma *MapAttr) clearOwner() {
	ma.owner = nil
	ma.parent = nil
	ma.pkey = nil
}

func NewMapAttr() *MapAttr {
	return &MapAttr{
		attrs: make(map[string]interface{}),
	}
}
