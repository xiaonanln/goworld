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

func (la *ListAttr) Set(index int, val interface{}) {
	la.items[index] = val
	if subma, ok := val.(*MapAttr); ok {
		// val is ListAttr, set parent and owner accordingly
		if subma.parent != nil || subma.owner != nil || subma.pkey != nil {
			gwlog.Panicf("ListAttr reused in index %s", index)
		}

		subma.parent = la
		subma.owner = la.owner
		subma.pkey = index

		la.sendListAttrChangeToClients(index, subma.ToMap())
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

func (la *ListAttr) getPathFromOwner() []string {
	path := make([]string, 0, 4) // preallocate some Space
	for {
		if la.parent != nil {
			path = append(path, la.pkey)
			la = la.parent
		} else { // la.parent  == nil, must be the root attr
			if la != la.owner.Attrs {
				gwlog.Panicf("Root attrs is not found")
			}
			break
		}
	}
	return path
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

func (la *ListAttr) GetStr(key string) string {
	val := la.Get(key)
	return val.(string)
}

func (la *ListAttr) GetFloat(key string) float64 {
	val := la.Get(key)
	return val.(float64)
}

func (la *ListAttr) GetBool(key string) bool {
	val := la.Get(key)
	return val.(bool)
}

func (la *ListAttr) GetListAttr(key string) *ListAttr {
	val := la.Get(key)
	return val.(*ListAttr)
}

// Delete a key in attrs
func (la *ListAttr) Pop(key string) interface{} {
	val, ok := la.items[key]
	if !ok {
		gwlog.Panicf("key not exists: %s", key)
	}

	delete(la.items, key)
	if subma, ok := val.(*ListAttr); ok {
		subma.parent = nil // clear parent and owner when attribute is poped
		subma.pkey = ""
		subma.owner = nil
	}

	la.sendAttrDelToClients(key)
	return val
}

func (la *ListAttr) Del(key string) {
	la.Pop(key)
}

func (la *ListAttr) PopListAttr(key string) *ListAttr {
	val := la.Pop(key)
	return val.(*ListAttr)
}

func (la *ListAttr) GetKeys() []string {
	size := len(la.items)
	keys := make([]string, 0, size)
	for k, _ := range la.items {
		keys = append(keys, k)
	}
	return keys
}

func (la *ListAttr) GetValues() []interface{} {
	size := len(la.items)
	vals := make([]interface{}, 0, size)
	for _, v := range la.items {
		vals = append(vals, v)
	}
	return vals
}

func (la *ListAttr) ToMap() map[string]interface{} {
	doc := map[string]interface{}{}
	for k, v := range la.items {
		innerListAttr, isInnerListAttr := v.(*ListAttr)
		if isInnerListAttr {
			doc[k] = innerListAttr.ToMap()
		} else {
			doc[k] = v
		}
	}
	return doc
}

func (la *ListAttr) ToMapWithFilter(filter func(string) bool) map[string]interface{} {
	doc := map[string]interface{}{}
	for k, v := range la.items {
		if !filter(k) {
			continue
		}

		innerListAttr, isInnerListAttr := v.(*ListAttr)
		if isInnerListAttr {
			doc[k] = innerListAttr.ToMap()
		} else {
			doc[k] = v
		}
	}
	return doc
}

func (la *ListAttr) AssignMap(doc map[string]interface{}) *ListAttr {
	for k, v := range doc {
		innerMap, ok := v.(map[string]interface{})
		if ok {
			innerListAttr := NewListAttr()
			innerListAttr.AssignMap(innerMap)
			la.Set(k, innerListAttr)
		} else {
			la.Set(k, v)
		}
	}
	return la
}

func (la *ListAttr) AssignMapWithFilter(doc map[string]interface{}, filter func(string) bool) *ListAttr {
	for k, v := range doc {
		if !filter(k) {
			continue
		}

		innerMap, ok := v.(map[string]interface{})
		if ok {
			innerListAttr := NewListAttr()
			innerListAttr.AssignMap(innerMap)
			la.Set(k, innerListAttr)
		} else {
			la.Set(k, v)
		}
	}
	return la
}

func NewListAttr() *ListAttr {
	return &ListAttr{
		items: []interface{}{},
	}
}
