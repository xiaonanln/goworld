package entity_manager

import . "github.com/xiaonanln/goworld/entity"

var (
	entityManager = newEntityManager()
)

type EntityManager struct {
	entities map[EntityID]*Entity
}

func newEntityManager() *EntityManager {
	return &EntityManager{
		entities: map[EntityID]*Entity{},
	}
}

func (em *EntityManager) Put(entity *Entity) {
	em.entities[entity.ID] = entity
}

func (em *EntityManager) Get(id EntityID) *Entity {
	return em.entities[id]
}
