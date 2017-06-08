package entity

import "bytes"
import . "github.com/xiaonanln/goworld/common"

type EntityMap map[EntityID]*Entity

func (em EntityMap) Add(entity *Entity) {
	em[entity.ID] = entity
}

func (em EntityMap) Del(id EntityID) {
	delete(em, id)
}

func (em EntityMap) Get(id EntityID) *Entity {
	return em[id]
}

type EntitySet map[*Entity]struct{}

func (es EntitySet) Add(entity *Entity) {
	es[entity] = struct{}{}
}

func (es EntitySet) Del(entity *Entity) {
	delete(es, entity)
}

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

type EntityIDSet map[EntityID]struct{}

func (es EntityIDSet) Add(id EntityID) {
	es[id] = struct{}{}
}

func (es EntityIDSet) Del(id EntityID) {
	delete(es, id)
}

func (es EntityIDSet) Contains(id EntityID) bool {
	_, ok := es[id]
	return ok
}
