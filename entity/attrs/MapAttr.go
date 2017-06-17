package attrs

import (
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/gwlog"
)

type MapAttr struct {
	attrs     map[string]interface{}
	dirtyKeys common.StringSet
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
	ma.dirtyKeys.Add(key)
}

func (ma *MapAttr) SetDefault(key string, val interface{}) {
	if _, ok := ma.attrs[key]; !ok {
		ma.attrs[key] = val
		ma.dirtyKeys.Add(key)
	}
}

func (ma *MapAttr) GetInt(key string) (iv int) {
	val, ok := ma.attrs[key]
	if !ok {
		gwlog.Panicf("key not exists: %s", key)
	}

	defer func() {
		recover()
		if i64, ok := val.(int64); ok {
			iv = int(i64)
			ma.attrs[key] = iv // fall back to int
			return
		}

		if f, ok := val.(float64); ok {
			iv = int(f)
			ma.attrs[key] = iv // fall back to int
			return
		}

		iv = val.(int) // will panic here
	}()

	iv = val.(int)
	return
}

func (ma *MapAttr) GetStr(key string) string {
	val, ok := ma.attrs[key]
	if !ok {
		gwlog.Panicf("key not exists: %s", key)
	}
	return val.(string)
}

func (ma *MapAttr) GetMapAttr(key string) *MapAttr {
	val, ok := ma.attrs[key]
	ma.dirtyKeys.Add(key)

	if !ok {
		attrs := NewMapAttr()
		ma.attrs[key] = attrs
		return attrs
	}
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

func (ma *MapAttr) GetMap() map[string]interface{} {
	return ma.attrs
}

func (ma *MapAttr) GetFloat(key string) float64 {
	val, ok := ma.attrs[key]
	if !ok {
		gwlog.Panicf("key not exists: %s", key)
	}
	return val.(float64)
}

func (ma *MapAttr) GetBool(key string) bool {
	val, ok := ma.attrs[key]
	if !ok {
		gwlog.Panicf("key not exists: %s", key)
	}
	return val.(bool)
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
			ma.attrs[k] = innerMapAttr
		} else {
			ma.attrs[k] = v
		}
	}
	return ma
}

func NewMapAttr() *MapAttr {
	return &MapAttr{
		attrs:     make(map[string]interface{}),
		dirtyKeys: common.StringSet{},
	}
}
