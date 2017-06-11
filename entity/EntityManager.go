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
	entityType2RpcDescMap = map[string]RpcDescMap{}
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
	entityType2RpcDescMap[typeName] = RpcDescMap{}

	entityPtrType := reflect.PtrTo(entityType)
	numMethods := entityPtrType.NumMethod()
	for i := 0; i < numMethods; i++ {
		method := entityPtrType.Method(i)
		entityType2RpcDescMap[typeName].visit(method)
	}

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
	entity.IV = entityPtrVal
	entity.I = entityPtrVal.Interface().(IEntity)
	entity.TypeName = typeName
	entity.rpcDescMap = entityType2RpcDescMap[typeName]

	entity.timers = map[*timer.Timer]struct{}{}
	initAOI(&entity.aoi)
	entity.I.OnInit()

	entityManager.put(entity)
	entity.Save() // save immediately after creation
	entity.I.OnCreated()

	//dispatcher_client.GetDispatcherClientForSend().SendNotifyCreateEntity(entityID)

	if space != nil {
		space.enter(entity)
	}

	return entityID
}

func createEntityAnywhere(typeName string) {
	dispatcher_client.GetDispatcherClientForSend().SendCreateEntityAnywhere(typeName)
}

func CreateEntityLocally(typeName string) EntityID {
	return createEntity(typeName, nil)
}

func CreateEntityAnywhere(typeName string) {
	createEntityAnywhere(typeName)
}

func callRemote(id EntityID, method string, args []interface{}) {
	gwlog.Info("dispatcher_client.GetDispatcherClientForSend(): %v", dispatcher_client.GetDispatcherClientForSend())
	dispatcher_client.GetDispatcherClientForSend().SendCallEntityMethod(id, method, args)
}

func OnCall(id EntityID, method string, args []interface{}) {
	e := entityManager.get(id)
	if e == nil {
		// entity not found, may destroyed before call
		gwlog.Warn("Entity %s is not found while calling %s%v", id, method, args)
		return
	}

	e.onCall(method, args)
}
