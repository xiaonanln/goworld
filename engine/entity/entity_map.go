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

// EntityIDSet is the data structure for a set of entity IDs
type EntityIDSet map[common.EntityID]struct{}

// Add adds an entity ID to EntityIDSet
func (es EntityIDSet) Add(id common.EntityID) {
	es[id] = struct{}{}
}

// Del removes an entity ID from EntityIDSet
func (es EntityIDSet) Del(id common.EntityID) {
	delete(es, id)
}

// Contains checks if entity ID is in EntityIDSet
func (es EntityIDSet) Contains(id common.EntityID) bool {
	_, ok := es[id]
	return ok
}

// ToList convert EntityIDSet to a slice of entity IDs
func (es EntityIDSet) ToList() []common.EntityID {
	list := make([]common.EntityID, 0, len(es))
	for eid := range es {
		list = append(list, eid)
	}
	return list
}
