package entity

import (
	"bytes"

	"github.com/xiaonanln/goworld/engine/common"
)

// EntityMap is the data structure for maintaining entity IDs to entities
type EntityMap map[common.EntityID]*Entity

// Add adds a new entity to EntityMap
func (em EntityMap) Add(entity *Entity) {
	em[entity.ID] = entity
}

// Del deletes an entity from EntityMap
func (em EntityMap) Del(id common.EntityID) {
	delete(em, id)
}

// Get returns the Entity of specified entity ID in EntityMap
func (em EntityMap) Get(id common.EntityID) *Entity {
	return em[id]
}

// Keys return keys of the EntityMap in a slice
func (em EntityMap) Keys() (keys []common.EntityID) {
	for eid := range em {
		keys = append(keys, eid)
	}
	return
}

// Values return values of the EntityMap in a slice
func (em EntityMap) Values() (vals []*Entity) {
	for _, e := range em {
		vals = append(vals, e)
	}
	return
}

// EntitySet is the data structure for a set of entities
type EntitySet map[*Entity]struct{}

// Add adds an entity to the EntitySet
func (es EntitySet) Add(entity *Entity) {
	es[entity] = struct{}{}
}

// Del deletes an entity from the EntitySet
func (es EntitySet) Del(entity *Entity) {
	delete(es, entity)
}

// Contains returns if the entity is in the EntitySet
func (es EntitySet) Contains(entity *Entity) bool {
	_, ok := es[entity]
	return ok
}

func (es EntitySet) ForEach(f func(e *Entity)) {
	for e := range es {
		f(e)
	}
}

func (es EntitySet) String() string {
	b := bytes.Buffer{}
	b.WriteString("{")
	first := true
	for entity := range es {
		if !first {
			b.WriteString(", ")
		} else {
			first = false
		}
		b.WriteString(entity.String())
	}
	b.WriteString("}")
	return b.String()
}
