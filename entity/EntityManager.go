package entity

import (
	"reflect"

	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/storage"
)

var (
	registeredEntityTypes = map[string]reflect.Type{}
	entityType2RpcDescMap = map[string]RpcDescMap{}
	entityManager         = newEntityManager()
)

type EntityManager struct {
	entities      EntityMap
	ownerOfClient map[ClientID]EntityID
}

func newEntityManager() *EntityManager {
	return &EntityManager{
		entities:      EntityMap{},
		ownerOfClient: map[ClientID]EntityID{},
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

func (em *EntityManager) onClientLoseOwner(clientid ClientID) {
	delete(em.ownerOfClient, clientid)
}

func (em *EntityManager) onClientSetOwner(clientid ClientID, eid EntityID) {
	em.ownerOfClient[clientid] = eid
}

func (em *EntityManager) onClientDisconnected(clientid ClientID) {
	eid := em.ownerOfClient[clientid]
	delete(em.ownerOfClient, clientid)

	if !eid.IsNil() { // should always true
		owner := em.entities[eid]
		owner.notifyClientDisconnected()
	}
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

func createEntity(typeName string, space *Space, entityID EntityID, data map[string]interface{}, client *GameClient) EntityID {
	gwlog.Debug("createEntity: %s in space %s", typeName, space)
	entityType, ok := registeredEntityTypes[typeName]
	if !ok {
		gwlog.Panicf("unknown entity type: %s", typeName)
	}

	if entityID == "" {
		entityID = GenEntityID()
	}

	entityPtrVal := reflect.New(entityType)
	entity := reflect.Indirect(entityPtrVal).FieldByName("Entity").Addr().Interface().(*Entity)
	entity.init(typeName, entityID, entityPtrVal)
	entity.space = nilSpace

	entityManager.put(entity)
	if data != nil {
		entity.LoadPersistentData(data)
	} else {
		entity.Save() // save immediately after creation
	}

	if entity.I.IsPersistent() { // startup the periodical timer for saving entity
		entity.setupSaveTimer()
		dispatcher_client.GetDispatcherClientForSend().SendNotifyCreateEntity(entityID)
	}

	entity.I.OnCreated()

	if client != nil {
		// assign client to the newly created entity
		entity.SetClient(client)
	}

	if space != nil {
		space.enter(entity)
	}

	return entityID
}

func loadEntityLocally(typeName string, entityID EntityID, space *Space) {
	// load the data from storage
	storage.Load(typeName, entityID, func(data interface{}, err error) {
		// callback runs in main routine
		if err != nil {
			gwlog.Panicf("load entity %s.%s failed: %s", typeName, entityID, err)
		}

		if space != nil && space.IsDestroyed() {
			// space might be destroy during the Load process, so cancel the entity creation
			return
		}

		createEntity(typeName, space, entityID, data.(map[string]interface{}), nil)
	})
}

func loadEntityAnywhere(typeName string, entityID EntityID) {
	dispatcher_client.GetDispatcherClientForSend().SendLoadEntityAnywhere(typeName, entityID)
}

func createEntityAnywhere(typeName string) {
	dispatcher_client.GetDispatcherClientForSend().SendCreateEntityAnywhere(typeName)
}

func CreateEntityLocally(typeName string, client *GameClient) EntityID {
	return createEntity(typeName, nil, "", nil, client)
}

func CreateEntityAnywhere(typeName string) {
	createEntityAnywhere(typeName)
}

func LoadEntityLocally(typeName string, entityID EntityID) {
	loadEntityLocally(typeName, entityID, nil)
}

func LoadEntityAnywhere(typeName string, entityID EntityID) {
	loadEntityAnywhere(typeName, entityID)
}

func OnClientDisconnected(clientid ClientID) {
	entityManager.onClientDisconnected(clientid) // pop the owner eid
}

func callRemote(id EntityID, method string, args []interface{}) {
	//gwlog.Info("dispatcher_client.GetDispatcherClientForSend(): %v", dispatcher_client.GetDispatcherClientForSend())
	dispatcher_client.GetDispatcherClientForSend().SendCallEntityMethod(id, method, args)
}

func OnCall(id EntityID, method string, args []interface{}, clientID ClientID) {
	e := entityManager.get(id)
	if e == nil {
		// entity not found, may destroyed before call
		gwlog.Warn("Entity %s is not found while calling %s%v", id, method, args)
		return
	}

	e.onCall(method, args, clientID)
}

func GetEntity(id EntityID) *Entity {
	return entityManager.get(id)
}
