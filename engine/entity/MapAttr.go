package entity

import (
	"strings"

	"fmt"

	"github.com/xiaonanln/goworld/engine/gwlog"
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

// String convert MapAttr to readable string
func (a *MapAttr) String() string {
	var sb strings.Builder
	sb.WriteString("MapAttr{")
	isFirstField := true
	for k, v := range a.attrs {
		if !isFirstField {
			sb.WriteString(", ")
		}

		fmt.Fprintf(&sb, "%#v", k)
		sb.WriteString(": ")
		switch a := v.(type) {
		case *MapAttr:
			sb.WriteString(a.String())
		case *ListAttr:
			sb.WriteString(a.String())
		default:
			fmt.Fprintf(&sb, "%#v", v)
		}
		isFirstField = false
	}
	sb.WriteString("}")
	return sb.String()
}

// HasKey returns if the key exists in MapAttr
func (a *MapAttr) HasKey(key string) bool {
	_, ok := a.attrs[key]
	return ok
}

// Keys returns all keys of Attrs
func (a *MapAttr) Keys() []string {
	keys := make([]string, 0, len(a.attrs))
	for k := range a.attrs {
		keys = append(keys, k)
	}
	return keys
}

// ForEachKey calls f on all keys
func (a *MapAttr) ForEachKey(f func(key string)) {
	for k := range a.attrs {
		f(k)
	}
}

// ForEach calls f on all items
// Be careful about the type of val
func (a *MapAttr) ForEach(f func(key string, val interface{})) {
	for k, v := range a.attrs {
		f(k, v)
	}
}

// Set sets the key-attribute pair in MapAttr
func (a *MapAttr) set(key string, val interface{}) {
	var flag attrFlag
	a.attrs[key] = val
	switch sa := val.(type) {
	case *MapAttr:
		// val is MapAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in key %s", key)
		}

		if a.owner != nil && a == a.owner.Attrs { // this is the root
			flag = a.owner.getAttrFlag(key)
		} else {
			flag = a.flag
		}
		sa.setParent(a.owner, a, key, flag)
		a.sendAttrChangeToClients(key, sa.ToMap())
	case *ListAttr:
		// val is ListATtr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("ListAttr reused in key %s", key)
		}

		if a.owner != nil && a == a.owner.Attrs { // this is the root
			flag = a.owner.getAttrFlag(key)
		} else {
			flag = a.flag
		}
		sa.setParent(a.owner, a, key, flag)
		a.sendAttrChangeToClients(key, sa.ToList())
	default:
		a.sendAttrChangeToClients(key, val)
	}
}

// SetInt sets int value at the key
func (a *MapAttr) SetInt(key string, v int64) {
	a.set(key, v)
}

// SetFloat sets float value at the key
func (a *MapAttr) SetFloat(key string, v float64) {
	a.set(key, v)
}

// SetBool sets bool value at the key
func (a *MapAttr) SetBool(key string, v bool) {
	a.set(key, v)
}

// SetStr sets string value at the key
func (a *MapAttr) SetStr(key string, v string) {
	a.set(key, v)
}

// SetMapAttr sets MapAttr value at the key
func (a *MapAttr) SetMapAttr(key string, attr *MapAttr) {
	a.set(key, attr)
}

// SetListAttr sets ListAttr value at the key
func (a *MapAttr) SetListAttr(key string, attr *ListAttr) {
	a.set(key, attr)
}

// SetDefaultInt sets default int value at the key
func (a *MapAttr) SetDefaultInt(key string, v int64) {
	if _, ok := a.attrs[key]; !ok {
		a.set(key, v)
	}
}

// SetDefaultFloat sets default float value at the key
func (a *MapAttr) SetDefaultFloat(key string, v float64) {
	if _, ok := a.attrs[key]; !ok {
		a.set(key, v)
	}
}

// SetDefaultBool sets default bool value at the key
func (a *MapAttr) SetDefaultBool(key string, v bool) {
	if _, ok := a.attrs[key]; !ok {
		a.set(key, v)
	}
}

// SetDefaultStr sets default string value at the key
func (a *MapAttr) SetDefaultStr(key string, v string) {
	if _, ok := a.attrs[key]; !ok {
		a.set(key, v)
	}
}

// SetDefaultMapAttr sets default MapAttr value at the key
func (a *MapAttr) SetDefaultMapAttr(key string, attr *MapAttr) {
	if _, ok := a.attrs[key]; !ok {
		a.set(key, attr)
	}
}

// SetDefaultListAttr sets default ListAttr value at the key
func (a *MapAttr) SetDefaultListAttr(key string, attr *ListAttr) {
	if _, ok := a.attrs[key]; !ok {
		a.set(key, attr)
	}
}

func (a *MapAttr) sendAttrChangeToClients(key string, val interface{}) {
	if a.owner != nil {
		// send the change to owner's Client
		a.owner.sendMapAttrChangeToClients(a, key, val)
	}
}

func (a *MapAttr) sendAttrDelToClients(key string) {
	if a.owner != nil {
		a.owner.sendMapAttrDelToClients(a, key)
	}
}

func (a *MapAttr) sendAttrClearToClients() {
	if a.owner != nil {
		a.owner.sendMapAttrClearToClients(a)
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

// get returns the attribute of specified key in MapAttr
func (a *MapAttr) get(key string) interface{} {
	val, ok := a.attrs[key]
	if !ok {
		gwlog.Panicf("key not exists: %s", key)
	}
	return val
}

// GetInt returns the attribute of specified key in MapAttr as int64
func (a *MapAttr) GetInt(key string) int64 {
	if val, ok := a.attrs[key]; ok {
		return val.(int64)
	} else {
		return 0
	}
}

// GetStr returns the attribute of specified key in MapAttr as string
func (a *MapAttr) GetStr(key string) string {
	if val, ok := a.attrs[key]; ok {
		return val.(string)
	} else {
		return ""
	}
}

// GetFloat returns the attribute of specified key in MapAttr as float64
func (a *MapAttr) GetFloat(key string) float64 {
	if val, ok := a.attrs[key]; ok {
		return val.(float64)
	} else {
		return 0
	}
}

// GetBool returns the attribute of specified key in MapAttr as bool
func (a *MapAttr) GetBool(key string) bool {
	if val, ok := a.attrs[key]; ok {
		return val.(bool)
	} else {
		return false
	}
}

// GetMapAttr returns the attribute of specified key in MapAttr as MapAttr
func (a *MapAttr) GetMapAttr(key string) *MapAttr {
	if val, ok := a.attrs[key]; ok {
		return val.(*MapAttr)
	} else {
		v := NewMapAttr()
		a.set(key, v)
		return v
	}
}

// GetListAttr returns the attribute of specified key in MapAttr as ListAttr
func (a *MapAttr) GetListAttr(key string) *ListAttr {
	if val, ok := a.attrs[key]; ok {
		return val.(*ListAttr)
	} else {
		v := NewListAttr()
		a.set(key, v)
		return v
	}
}

// Pop deletes a key in MapAttr and returns the attribute
func (a *MapAttr) pop(key string) interface{} {
	val, ok := a.attrs[key]
	if !ok {
		return nil
	}

	delete(a.attrs, key)
	switch sa := val.(type) {
	case *MapAttr:
		sa.removeFromParent()
	case *ListAttr:
		sa.removeFromParent()
	}

	a.sendAttrDelToClients(key)
	return val
}

// Del deletes a key in MapAttr
func (a *MapAttr) Del(key string) {
	a.pop(key)
}

// PopInt deletes a key in MapAttr and returns the attribute as int64
func (a *MapAttr) PopInt(key string) int64 {
	val := a.pop(key)
	if val != nil {
		return val.(int64)
	} else {
		return 0
	}
}

// PopFloat deletes a key in MapAttr and returns the attribute as float64
func (a *MapAttr) PopFloat(key string) float64 {
	val := a.pop(key)
	if val != nil {
		return val.(float64)
	} else {
		return 0.0
	}
}

// PopBool deletes a key in MapAttr and returns the attribute as bool
func (a *MapAttr) PopBool(key string) bool {
	val := a.pop(key)
	if val != nil {
		return val.(bool)
	} else {
		return false
	}
}

// PopStr deletes a key in MapAttr and returns the attribute as str
func (a *MapAttr) PopStr(key string) string {
	val := a.pop(key)
	if val != nil {
		return val.(string)
	} else {
		return ""
	}
}

// PopMapAttr deletes a key in MapAttr and returns the attribute as MapAttr
func (a *MapAttr) PopMapAttr(key string) *MapAttr {
	val := a.pop(key)
	if val != nil {
		return val.(*MapAttr)
	} else {
		return nil
	}
}

// PopListAttr deletes a key in MapAttr and returns the attribute as MapAttr
func (a *MapAttr) PopListAttr(key string) *ListAttr {
	val := a.pop(key)
	if val != nil {
		return val.(*ListAttr)
	} else {
		return nil
	}
}

// Clear removes all key-values from the MapAttr
func (a *MapAttr) Clear() {
	if len(a.attrs) == 0 {
		return
	}

	var curattrs map[string]interface{}
	curattrs, a.attrs = a.attrs, map[string]interface{}{}
	for _, v := range curattrs {
		switch sa := v.(type) {
		case *MapAttr:
			sa.removeFromParent()
		case *ListAttr:
			sa.removeFromParent()
		}
	}

	a.sendAttrClearToClients()
}

// ToMap converts MapAttr to native map, recursively
func (a *MapAttr) ToMap() map[string]interface{} {
	doc := map[string]interface{}{}
	for k, v := range a.attrs {
		switch a := v.(type) {
		case *MapAttr:
			doc[k] = a.ToMap()
		case *ListAttr:
			doc[k] = a.ToList()
		default:
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

		switch a := v.(type) {
		case *MapAttr:
			doc[k] = a.ToMap()
		case *ListAttr:
			doc[k] = a.ToList()
		default:
			doc[k] = v
		}
	}
	return doc
}

// AssignMap assigns native map to MapAttr recursively
func (a *MapAttr) AssignMap(doc map[string]interface{}) {
	for k, v := range doc {
		switch iv := v.(type) {
		case map[string]interface{}:
			ia := NewMapAttr()
			ia.AssignMap(iv)
			a.set(k, ia)
		case []interface{}:
			ia := NewListAttr()
			ia.AssignList(iv)
			a.set(k, ia)
		default:
			a.set(k, uniformAttrType(v))
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
			a.set(k, ia)
		} else if iv, ok := v.([]interface{}); ok {
			ia := NewListAttr()
			ia.AssignList(iv)
			a.set(k, ia)
		} else {
			a.set(k, uniformAttrType(v))
		}
	}
}

func (a *MapAttr) removeFromParent() {
	a.parent = nil
	a.pkey = nil
	a.clearOwner()
}

func (a *MapAttr) clearOwner() {
	a.owner = nil
	a.path = nil
	a.flag = 0

	// clear owner of children recursively
	for _, v := range a.attrs {
		switch a := v.(type) {
		case *MapAttr:
			a.clearOwner()
		case *ListAttr:
			a.clearOwner()
		}
	}
}

func (a *MapAttr) setParent(owner *Entity, parent interface{}, pkey interface{}, flag attrFlag) {
	a.parent = parent
	a.pkey = pkey
	a.setOwner(owner, flag)
}

func (a *MapAttr) setOwner(owner *Entity, flag attrFlag) {
	a.owner = owner
	a.flag = flag

	// set owner of children recursively
	for _, v := range a.attrs {
		switch a := v.(type) {
		case *MapAttr:
			a.setOwner(owner, flag)
		case *ListAttr:
			a.setOwner(owner, flag)
		}
	}
}

// NewMapAttr creates a new MapAttr
func NewMapAttr() *MapAttr {
	return &MapAttr{
		attrs: make(map[string]interface{}),
	}
}
