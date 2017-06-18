package entity

import (
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/typeconv"
)

type MapAttr struct {
	owner  *Entity
	parent *MapAttr
	pkey   string // key of this item in parent
	attrs  map[string]interface{}
}

func (ma *MapAttr) Size() int {
	return len(ma.attrs)
}

func (ma *MapAttr) getOwner() *Entity {
	return ma.owner
}

func (ma *MapAttr) HasKey(key string) bool {
	_, ok := ma.attrs[key]
	return ok
}

func (ma *MapAttr) Set(key string, val interface{}) {
	ma.attrs[key] = val
	if subma, ok := val.(*MapAttr); ok {
		// val is MapAttr, set parent and owner accordingly
		if subma.parent != nil || subma.owner != nil || subma.pkey != "" {
			gwlog.Panicf("MapAttr reused in key %s", key)
		}

		subma.parent = ma
		subma.owner = ma.owner
		subma.pkey = key

		ma.sendAttrChangeToClients(key, subma.ToMap())
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
	owner := ma.getOwner()
	if owner != nil {
		// send the change to owner's client
		owner.sendAttrChangeToClients(ma, key, val)
	}
}

func (ma *MapAttr) sendAttrDelToClients(key string) {
	owner := ma.getOwner()
	if owner != nil {
		owner.sendAttrDelToClients(ma, key)
	}
}

func (ma *MapAttr) getPathFromOwner() []string {
	path := make([]string, 0, 4) // preallocate some Space
	for {
		if ma.parent != nil {
			path = append(path, ma.pkey)
			ma = ma.parent
		} else { // ma.parent  == nil, must be the root attr
			if ma != ma.owner.Attrs {
				gwlog.Panicf("Root attrs is not found")
			}
			break
		}
	}
	return path
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

// Delete a key in attrs
func (ma *MapAttr) Pop(key string) interface{} {
	val, ok := ma.attrs[key]
	if !ok {
		gwlog.Panicf("key not exists: %s", key)
	}

	delete(ma.attrs, key)
	if subma, ok := val.(*MapAttr); ok {
		subma.parent = nil // clear parent and owner when attribute is poped
		subma.pkey = ""
		subma.owner = nil
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
		innerMapAttr, isInnerMapAttr := v.(*MapAttr)
		if isInnerMapAttr {
			doc[k] = innerMapAttr.ToMap()
		} else {
			doc[k] = v
		}
	}
	return doc
}

func (ma *MapAttr) AssignMap(doc map[string]interface{}) *MapAttr {
	for k, v := range doc {
		innerMap, ok := v.(map[string]interface{})
		if ok {
			innerMapAttr := NewMapAttr()
			innerMapAttr.AssignMap(innerMap)
			ma.Set(k, innerMapAttr)
		} else {
			ma.Set(k, v)
		}
	}
	return ma
}

func NewMapAttr() *MapAttr {
	return &MapAttr{
		attrs: make(map[string]interface{}),
	}
}
