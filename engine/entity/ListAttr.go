package entity

import (
	"fmt"
	"strings"

	"github.com/xiaonanln/goworld/engine/gwlog"
)

// ListAttr is a attribute for a list of attributes
type ListAttr struct {
	owner  *Entity
	parent interface{}
	pkey   interface{} // key of this item in parent
	path   []interface{}
	flag   attrFlag
	items  []interface{}
}

func (a *ListAttr) String() string {
	var sb strings.Builder
	sb.WriteString("ListAttr{")
	isFirstField := true
	for _, v := range a.items {
		if !isFirstField {
			sb.WriteString(", ")
		}

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

// Size returns size of ListAttr
func (a *ListAttr) Size() int {
	return len(a.items)
}

func (a *ListAttr) removeFromParent() {
	a.parent = nil
	a.pkey = nil

	a.clearOwner()
}

func (a *ListAttr) clearOwner() {
	a.owner = nil
	a.flag = 0
	a.path = nil

	// clear owner of children recursively
	for _, v := range a.items {
		switch a := v.(type) {
		case *MapAttr:
			a.clearOwner()
		case *ListAttr:
			a.clearOwner()
		}
	}
}

func (a *ListAttr) setParent(owner *Entity, parent interface{}, pkey interface{}, flag attrFlag) {
	a.parent = parent
	a.pkey = pkey

	a.setOwner(owner, flag)
}

func (a *ListAttr) setOwner(owner *Entity, flag attrFlag) {
	a.owner = owner
	a.flag = flag

	// set owner of children recursively
	for _, v := range a.items {
		switch a := v.(type) {
		case *MapAttr:
			a.setOwner(owner, flag)
		case *ListAttr:
			a.setOwner(owner, flag)
		}
	}
}

// Set sets item value
func (a *ListAttr) set(index int, val interface{}) {
	a.items[index] = val
	switch sa := val.(type) {
	case *MapAttr:
		// val is MapAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in index %d", index)
		}

		sa.setParent(a.owner, a, index, a.flag)
		a.sendListAttrChangeToClients(index, sa.ToMap())
	case *ListAttr:
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in index %d", index)
		}

		sa.setParent(a.owner, a, index, a.flag)
		a.sendListAttrChangeToClients(index, sa.ToList())
	default:
		a.sendListAttrChangeToClients(index, val)
	}
}

func (a *ListAttr) sendListAttrChangeToClients(index int, val interface{}) {
	owner := a.owner
	if owner != nil {
		// send the change to owner's Client
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
	}
	return path
}

// Get gets item value
func (a *ListAttr) get(index int) interface{} {
	return a.items[index]
}

// GetInt gets item value as int
func (a *ListAttr) GetInt(index int) int64 {
	return a.get(index).(int64)
}

// GetFloat gets item value as float64
func (a *ListAttr) GetFloat(index int) float64 {
	return a.get(index).(float64)
}

// GetStr gets item value as string
func (a *ListAttr) GetStr(index int) string {
	return a.get(index).(string)
}

// GetBool gets item value as bool
func (a *ListAttr) GetBool(index int) bool {
	return a.get(index).(bool)
}

// GetListAttr gets item value as ListAttr
func (a *ListAttr) GetListAttr(index int) *ListAttr {
	val := a.get(index)
	return val.(*ListAttr)
}

// GetMapAttr gets item value as MapAttr
func (a *ListAttr) GetMapAttr(index int) *MapAttr {
	val := a.get(index)
	return val.(*MapAttr)
}

// AppendInt puts int value to the end of list
func (a *ListAttr) AppendInt(v int64) {
	a.append(v)
}

// AppendFloat puts float value to the end of list
func (a *ListAttr) AppendFloat(v float64) {
	a.append(v)
}

// AppendBool puts bool value to the end of list
func (a *ListAttr) AppendBool(v bool) {
	a.append(v)
}

// AppendStr puts string value to the end of list
func (a *ListAttr) AppendStr(v string) {
	a.append(v)
}

// AppendMapAttr puts MapAttr value to the end of list
func (a *ListAttr) AppendMapAttr(attr *MapAttr) {
	a.append(attr)
}

// AppendListAttr puts ListAttr value to the end of list
func (a *ListAttr) AppendListAttr(attr *ListAttr) {
	a.append(attr)
}

// Pop removes the last item from the end
func (a *ListAttr) pop() interface{} {
	size := len(a.items)
	val := a.items[size-1]
	a.items = a.items[:size-1]

	switch sa := val.(type) {
	case *MapAttr:
		sa.removeFromParent()
	case *ListAttr:
		sa.removeFromParent()
	}

	a.sendListAttrPopToClients()
	return val
}

func (a *ListAttr) PopInt() int64 {
	return a.pop().(int64)
}

func (a *ListAttr) PopFloat() float64 {
	return a.pop().(float64)
}

func (a *ListAttr) PopBool() bool {
	return a.pop().(bool)
}

func (a *ListAttr) PopStr() string {
	return a.pop().(string)
}

// PopListAttr removes the last item and returns as ListAttr
func (a *ListAttr) PopListAttr() *ListAttr {
	return a.pop().(*ListAttr)
}

// PopMapAttr removes the last item and returns as MapAttr
func (a *ListAttr) PopMapAttr() *MapAttr {
	return a.pop().(*MapAttr)
}

// append puts item to the end of list
func (a *ListAttr) append(val interface{}) {
	a.items = append(a.items, val)
	index := len(a.items) - 1

	switch sa := val.(type) {
	case *MapAttr:
		// val is ListAttr, set parent and owner accordingly
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in append")
		}

		sa.parent = a
		sa.owner = a.owner
		sa.pkey = index
		sa.flag = a.flag

		a.sendListAttrAppendToClients(sa.ToMap())
	case *ListAttr:
		if sa.parent != nil || sa.owner != nil || sa.pkey != nil {
			gwlog.Panicf("MapAttr reused in append")
		}

		sa.setParent(a.owner, a, index, a.flag)
		a.sendListAttrAppendToClients(sa.ToList())
	default:
		a.sendListAttrAppendToClients(val)
	}
}

// SetInt sets int value at the index
func (a *ListAttr) SetInt(index int, v int64) {
	a.set(index, v)
}

// SetFloat sets float value at the index
func (a *ListAttr) SetFloat(index int, v float64) {
	a.set(index, v)
}

// SetBool sets bool value at the index
func (a *ListAttr) SetBool(index int, v bool) {
	a.set(index, v)
}

// SetStr sets string value at the index
func (a *ListAttr) SetStr(index int, v string) {
	a.set(index, v)
}

// SetMapAttr sets MapAttr value at the index
func (a *ListAttr) SetMapAttr(index int, attr *MapAttr) {
	a.set(index, attr)
}

// SetListAttr sets ListAttr value at the index
func (a *ListAttr) SetListAttr(index int, attr *ListAttr) {
	a.set(index, attr)
}

// ToList converts ListAttr to slice, recursively
func (a *ListAttr) ToList() []interface{} {
	l := make([]interface{}, len(a.items))

	for i, v := range a.items {
		switch a := v.(type) {
		case *MapAttr:
			l[i] = a.ToMap()
		case *ListAttr:
			l[i] = a.ToList()
		default:
			l[i] = v
		}
	}
	return l
}

// AssignList assigns slice to ListAttr, recursively
func (a *ListAttr) AssignList(l []interface{}) {
	for _, v := range l {
		switch iv := v.(type) {
		case map[string]interface{}:
			ia := NewMapAttr()
			ia.AssignMap(iv)
			a.append(ia)
		case []interface{}:
			ia := NewListAttr()
			ia.AssignList(iv)
			a.append(ia)
		default:
			a.append(uniformAttrType(v))
		}
	}
}

// NewListAttr creates a new ListAttr
func NewListAttr() *ListAttr {
	return &ListAttr{
		items: []interface{}{},
	}
}
