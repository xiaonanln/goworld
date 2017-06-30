package entity

import (
	"reflect"

	"math/rand"

	"os"

	. "github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/gwutils"
	"github.com/xiaonanln/goworld/storage"
)

var (
	registeredEntityTypes = map[string]reflect.Type{}
	entityType2RpcDescMap = map[string]RpcDescMap{}
	entityManager         = newEntityManager()
)

type EntityManager struct {
	entities           EntityMap
	ownerOfClient      map[ClientID]EntityID
	registeredServices map[string]EntityIDSet
	filterTrees        map[string]*gwutils.FilterTree
}

func newEntityManager() *EntityManager {
	return &EntityManager{
		entities:           EntityMap{},
		ownerOfClient:      map[ClientID]EntityID{},
		registeredServices: map[string]EntityIDSet{},
		//filterTrees:        map[string]*gwutils.FilterTree{}, //  filter props only supported in clients
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

func (em *EntityManager) onEntityLoseClient(clientid ClientID) {
	delete(em.ownerOfClient, clientid)
}

func (em *EntityManager) onEntityGetClient(entityID EntityID, clientid ClientID) {
	em.ownerOfClient[clientid] = entityID
}

func (em *EntityManager) onClientDisconnected(clientid ClientID) {
	eid := em.ownerOfClient[clientid]
	if !eid.IsNil() { // should always true
		em.onEntityLoseClient(clientid)

		owner := em.get(eid) // FIXME: owner should not be nil

		if consts.DEBUG_CLIENTS {
			if owner == nil {
				gwlog.Warn("Client %s can not find owner entity %s", clientid, eid)
			}
		}

		owner.notifyClientDisconnected()
	}
}

func (em *EntityManager) onDeclareService(serviceName string, eid EntityID) {
	eids, ok := em.registeredServices[serviceName]
	if !ok {
		eids = EntityIDSet{}
		em.registeredServices[serviceName] = eids
	}
	eids.Add(eid)
}

func (em *EntityManager) onUndeclareService(serviceName string, eid EntityID) {
	eids, ok := em.registeredServices[serviceName]
	if ok {
		eids.Del(eid)
	}
}

func (em *EntityManager) chooseServiceProvider(serviceName string) EntityID {
	// choose one entity ID of service providers randomly
	eids, ok := em.registeredServices[serviceName]
	if !ok {
		gwlog.Panicf("service not found: %s", serviceName)
	}

	r := rand.Intn(len(eids)) // get a random one
	for eid := range eids {
		if r == 0 {
			return eid
		}
		r -= 1
	}
	return "" // never goes here
}

func RegisterEntity(typeName string, entityPtr IEntity) {
	if _, ok := registeredEntityTypes[typeName]; ok {
		gwlog.Panicf("RegisterEntity: Entity type %s already registered", typeName)
	}
	entityVal := reflect.Indirect(reflect.ValueOf(entityPtr))
	entityType := entityVal.Type()

	// register the string of e
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

func createEntity(typeName string, space *Space, entityID EntityID, data map[string]interface{}, client *GameClient, isMigrate bool) EntityID {
	//gwlog.Debug("createEntity: %s in Space %s", typeName, space)
	entityType, ok := registeredEntityTypes[typeName]
	if !ok {
		gwlog.Panicf("unknown entity type: %s", typeName)
		if consts.DEBUG_MODE {
			os.Exit(2)
		}
	}

	if entityID == "" {
		entityID = GenEntityID()
	}

	var entity *Entity
	var entityInstance reflect.Value

	//if typeName != SPACE_ENTITY_TYPE {
	entityInstance = reflect.New(entityType)
	//} else {
	//	entityInstance = reflect.New(spaceType)
	//}
	entity = reflect.Indirect(entityInstance).FieldByName("Entity").Addr().Interface().(*Entity)
	entity.init(typeName, entityID, entityInstance)
	entity.Space = nilSpace

	entityManager.put(entity)
	if data != nil {
		entity.I.LoadPersistentData(data)
	} else {
		entity.Save() // save immediately after creation
	}

	isPersistent := entity.I.IsPersistent()
	if isPersistent { // startup the periodical timer for saving e
		entity.setupSaveTimer()
	}

	if !isMigrate {
		dispatcher_client.GetDispatcherClientForSend().SendNotifyCreateEntity(entityID)
	}

	if client != nil {
		// assign client to the newly created
		if !isMigrate {
			entity.SetClient(client)
		} else {
			entity.client = client // assign client quietly if migrate
			entityManager.onEntityGetClient(entity.ID, client.clientid)
		}
	}

	if !isMigrate {
		gwutils.RunPanicless(entity.I.OnCreated)
	} else {
		gwutils.RunPanicless(entity.I.OnMigrateIn)
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
			// TODO: need to notify dispatcher that entity creation failed ?
			dispatcher_client.GetDispatcherClientForSend().SendNotifyDestroyEntity(entityID) // load entity failed, tell dispatcher
		}

		if space != nil && space.IsDestroyed() {
			// Space might be destroy during the Load process, so cancel the entity creation
			dispatcher_client.GetDispatcherClientForSend().SendNotifyDestroyEntity(entityID) // load entity failed, tell dispatcher
			return
		}

		createEntity(typeName, space, entityID, data.(map[string]interface{}), nil, false)
	})
}

func loadEntityAnywhere(typeName string, entityID EntityID) {
	dispatcher_client.GetDispatcherClientForSend().SendLoadEntityAnywhere(typeName, entityID)
}

func createEntityAnywhere(typeName string, data map[string]interface{}) {
	dispatcher_client.GetDispatcherClientForSend().SendCreateEntityAnywhere(typeName, data)
}

func CreateEntityLocally(typeName string, data map[string]interface{}, client *GameClient) EntityID {
	return createEntity(typeName, nil, "", data, client, false)
}

func CreateEntityAnywhere(typeName string) {
	createEntityAnywhere(typeName, nil)
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

func OnDeclareService(serviceName string, entityid EntityID) {
	entityManager.onDeclareService(serviceName, entityid)
}

func OnUndeclareService(serviceName string, entityid EntityID) {
	entityManager.onUndeclareService(serviceName, entityid)
}

func GetServiceProviders(serviceName string) EntityIDSet {
	return entityManager.registeredServices[serviceName]
}

func callRemote(id EntityID, method string, args []interface{}) {
	//gwlog.Info("dispatcher_client.GetDispatcherClientForSend(): %v", dispatcher_client.GetDispatcherClientForSend())
	dispatcher_client.GetDispatcherClientForSend().SendCallEntityMethod(id, method, args)
}

func OnCall(id EntityID, method string, args []interface{}, clientID ClientID) {
	e := entityManager.get(id)
	if e == nil {
		// entity not found, may destroyed before call
		gwlog.Error("Entity %s is not found while calling %s%v", id, method, args)
		return
	}

	e.onCall(method, args, clientID)
}

func GetEntity(id EntityID) *Entity {
	return entityManager.get(id)
}
