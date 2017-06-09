package entity

import (
	"reflect"

	timer "github.com/xiaonanln/goTimer"
	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/gwlog"
)

var (
	registeredEntityTypes = map[string]reflect.Type{}
	entityManager         = newEntityManager()
)

type EntityManager struct {
	entities EntityMap
}

func newEntityManager() *EntityManager {
	return &EntityManager{
		entities: EntityMap{},
	}
}

func (em *EntityManager) put(entity *Entity) {
	em.entities.Add(entity)
}

func (em *EntityManager) del(entityID EntityID) {
	em.entities.Del(entityID)
}

func (em *EntityManager) get(id EntityID) *Entity {
	return em.entities.Get(id)
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

	entity.timers = map[*timer.Timer]struct{}{}
	initAOI(&entity.aoi)
	entity.I.OnInit()

	entityManager.put(entity)
	entity.I.OnCreated()

	dispatcher_client.GetDispatcherClientForSend().SendNotifyCreateEntity(entityID)

	if space != nil {
		space.enter(entity)
	}

	return entityID
}

func CreateEntity(typeName string) EntityID {
	return createEntity(typeName, nil)
}

func call(id EntityID, method string, args []interface{}) {
	dispatcher_client.GetDispatcherClientForSend().SendCallEntityMethod(id, method)
}
