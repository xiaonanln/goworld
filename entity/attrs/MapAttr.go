package attrs

import "github.com/xiaonanln/goworld/common"

type MapAttr struct {
	attrs     map[string]interface{}
	dirtyKeys common.StringSet
}

func (self *MapAttr) Size() int {
	return len(self.attrs)
}

func (self *MapAttr) HasKey(key string) bool {
	_, ok := self.attrs[key]
	return ok
}

func (self *MapAttr) Set(key string, val interface{}) {
	self.attrs[key] = val
	self.dirtyKeys.Add(key)
}

func (self *MapAttr) SetDefault(key string, val interface{}) {
	if _, ok := self.attrs[key]; !ok {
		self.attrs[key] = val
		self.dirtyKeys.Add(key)
	}
}

func (self *MapAttr) GetInt(key string, defaultVal int) int {
	val, ok := self.attrs[key]
	if !ok {
		return defaultVal
	}
	i64, ok := val.(int64)
	if ok {
		return int(i64)
	}

	return val.(int)
}

func (self *MapAttr) GetStr(key string, defaultVal string) string {
	val, ok := self.attrs[key]
	if !ok {
		return defaultVal
	}
	return val.(string)
}

func (self *MapAttr) GetMapAttr(key string) *MapAttr {
	val, ok := self.attrs[key]
	self.dirtyKeys.Add(key)

	if !ok {
		attrs := NewMapAttr()
		self.attrs[key] = attrs
		return attrs
	}
	return val.(*MapAttr)
}

func (self *MapAttr) GetKeys() []string {
	size := len(self.attrs)
	keys := make([]string, 0, size)
	for k, _ := range self.attrs {
		keys = append(keys, k)
	}
	return keys
}

func (self *MapAttr) GetValues() []interface{} {
	size := len(self.attrs)
	vals := make([]interface{}, 0, size)
	for _, v := range self.attrs {
		vals = append(vals, v)
	}
	return vals
}

func (self *MapAttr) GetMap() map[string]interface{} {
	return self.attrs
}

func (self *MapAttr) GetFloat(key string, defaultVal float64) float64 {
	val, ok := self.attrs[key]
	if !ok {
		return defaultVal
	}
	return val.(float64)
}

func (self *MapAttr) GetBool(key string, defaultVal bool) bool {
	val, ok := self.attrs[key]
	if !ok {
		return defaultVal
	}
	return val.(bool)
}

func (self *MapAttr) ToMap() map[string]interface{} {
	doc := map[string]interface{}{}
	for k, v := range self.attrs {
		innerMapAttr, isInnerMapAttr := v.(*MapAttr)
		if isInnerMapAttr {
			doc[k] = innerMapAttr.ToMap()
		} else {
			doc[k] = v
		}
	}
	return doc
}

func (self *MapAttr) AssignMap(doc map[string]interface{}) *MapAttr {
	for k, v := range doc {
		innerMap, ok := v.(map[string]interface{})
		if ok {
			innerMapAttr := NewMapAttr()
			innerMapAttr.AssignMap(innerMap)
			self.attrs[k] = innerMapAttr
		} else {
			self.attrs[k] = v
		}
	}
	return self
}

func NewMapAttr() *MapAttr {
	return &MapAttr{
		attrs:     make(map[string]interface{}),
		dirtyKeys: common.StringSet{},
	}
}
