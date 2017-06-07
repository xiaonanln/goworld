package entity

import (
	"reflect"

	"github.com/xiaonanln/goworld/gwlog"
)

var (
	registeredEntityTypes = map[string]reflect.Type{}
	entityManager         = newEntityManager()
)

type EntityManager struct {
	entities map[EntityID]IEntity
}

func newEntityManager() *EntityManager {
	return &EntityManager{
		entities: map[EntityID]IEntity{},
	}
}

func (em *EntityManager) Put(entity *Entity) {
	em.entities[entity.ID] = entity
}

func (em *EntityManager) Get(id EntityID) IEntity {
	return em.entities[id]
}

func RegisterEntity(typeName string, entityPtr IEntity) {
	if _, ok := registeredEntityTypes[typeName]; ok {
		gwlog.Panicf("RegisterEntity: Entity type %s already registered", typeName)
	}
	entityVal := reflect.Indirect(reflect.ValueOf(entityPtr))
	entityType := entityVal.Type()

	// register the string of entity
	registeredEntityTypes[typeName] = entityType

	gwlog.Debug(">>> RegisterEntity %s => %s <<<", typeName, entityType.Name())
}

func createEntity(typeName string, space *Space) EntityID {
	gwlog.Debug("createEntity: %s in space %s", typeName, space)
	entityType, ok := registeredEntityTypes[typeName]
	if !ok {
		gwlog.Panicf("unknown entity type: %s", typeName)
	}

	entityID := GenEntityID()
	entityPtrVal := reflect.New(entityType)
	entity := reflect.Indirect(entityPtrVal).FieldByName("Entity").Addr().Interface().(*Entity)
	entity.ID = entityID
	entity.I = entityPtrVal.Interface().(IEntity)
	entity.TypeName = typeName
	entityManager.Put(entity)
	entity.I.OnCreated()

	if space != nil {
		space.enter(entity)
	}

	return entityID
}
