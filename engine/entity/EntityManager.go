package entity

import (
	"reflect"

	"math/rand"

	"strings"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/dispatchercluster"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/goworld/engine/storage"
	"github.com/xiaonanln/typeconv"
)

var (
	registeredEntityTypes = map[string]*EntityTypeDesc{}
	entityManager         = newEntityManager()
)

// EntityTypeDesc is the entity type description for registering entity types
type EntityTypeDesc struct {
	isPersistent    bool
	useAOI          bool
	entityType      reflect.Type
	rpcDescs        rpcDescMap
	allClientAttrs  common.StringSet
	clientAttrs     common.StringSet
	persistentAttrs common.StringSet
	isService       bool
	//compositiveMethodComponentIndices map[string][]int
	//definedAttrs                      bool
}

var _VALID_ATTR_DEFS = common.StringSet{} // all valid attribute defs

func init() {
	_VALID_ATTR_DEFS.Add(strings.ToLower("Client"))
	_VALID_ATTR_DEFS.Add(strings.ToLower("AllClients"))
	_VALID_ATTR_DEFS.Add(strings.ToLower("Persistent"))
}

func (desc *EntityTypeDesc) SetPersistent(persistent bool) *EntityTypeDesc {
	desc.isPersistent = persistent
	return desc
}

func (desc *EntityTypeDesc) SetUseAOI(useAOI bool) *EntityTypeDesc {
	desc.useAOI = useAOI
	return desc
}

func (desc *EntityTypeDesc) DefineAttr(attr string, defs ...string) *EntityTypeDesc {
	gwlog.Infof("        Attr %s = %v", attr, defs)
	isAllClient, isClient, isPersistent := false, false, false

	for _, def := range defs {
		def := strings.ToLower(def)

		if !_VALID_ATTR_DEFS.Contains(def) {
			// not a valid def
			gwlog.Panicf("attribute %s: invalid property: %s; all valid properties: %v", attr, def, _VALID_ATTR_DEFS.ToList())
		}

		if def == "allclients" {
			isAllClient = true
			isClient = true
		} else if def == "client" {
			isClient = true
		} else if def == "persistent" {
			isPersistent = true
			// make sure non-persistent entity has no persistent attributes
			if !desc.isPersistent {
				gwlog.Fatalf("Entity type %s is not persistent, should not define persistent attribute: %s", desc.entityType.Name(), attr)
			}
		}
	}

	if isAllClient {
		desc.allClientAttrs.Add(attr)
	}
	if isClient {
		desc.clientAttrs.Add(attr)
	}
	if isPersistent {
		desc.persistentAttrs.Add(attr)
	}
	return desc
}

type _EntityManager struct {
	entities           EntityMap
	ownerOfClient      map[common.ClientID]common.EntityID
	registeredServices map[string]EntityIDSet
}

func newEntityManager() *_EntityManager {
	return &_EntityManager{
		entities:           EntityMap{},
		ownerOfClient:      map[common.ClientID]common.EntityID{},
		registeredServices: map[string]EntityIDSet{},
	}
}

func (em *_EntityManager) put(entity *Entity) {
	em.entities.Add(entity)
}

func (em *_EntityManager) del(entityID common.EntityID) {
	em.entities.Del(entityID)
}

func (em *_EntityManager) get(id common.EntityID) *Entity {
	return em.entities.Get(id)
}

func (em *_EntityManager) onEntityLoseClient(clientid common.ClientID) {
	delete(em.ownerOfClient, clientid)
}

func (em *_EntityManager) onEntityGetClient(entityID common.EntityID, clientid common.ClientID) {
	em.ownerOfClient[clientid] = entityID
}

func (em *_EntityManager) onClientDisconnected(clientid common.ClientID) {
	eid := em.ownerOfClient[clientid]
	if !eid.IsNil() { // should always true
		em.onEntityLoseClient(clientid)
		owner := em.get(eid)
		owner.notifyClientDisconnected()
	}
}

func (em *_EntityManager) onGateDisconnected(gateid uint16) {
	for _, entity := range em.entities {
		client := entity.client
		if client != nil && client.gateid == gateid {
			em.onEntityLoseClient(client.clientid)
			entity.notifyClientDisconnected()
		}
	}
}

func (em *_EntityManager) onDeclareService(serviceName string, eid common.EntityID) {
	eids, ok := em.registeredServices[serviceName]
	if !ok {
		eids = EntityIDSet{}
		em.registeredServices[serviceName] = eids
	}
	eids.Add(eid)
}

func (em *_EntityManager) onUndeclareService(serviceName string, eid common.EntityID) {
	eids, ok := em.registeredServices[serviceName]
	if ok {
		eids.Del(eid)
	}
}

func (em *_EntityManager) chooseServiceProvider(serviceName string) common.EntityID {
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

// RegisterEntity registers custom entity type and define entity behaviors
func RegisterEntity(typeName string, entity IEntity, isService bool) *EntityTypeDesc {
	if _, ok := registeredEntityTypes[typeName]; ok {
		gwlog.Panicf("RegisterEntity: Entity type %s already registered", typeName)
	}

	entityVal := reflect.ValueOf(entity)
	entityType := entityVal.Type()

	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	// register the string of e
	rpcDescs := rpcDescMap{}
	entityTypeDesc := &EntityTypeDesc{
		isPersistent:    false,
		useAOI:          false,
		entityType:      entityType,
		rpcDescs:        rpcDescs,
		clientAttrs:     common.StringSet{},
		allClientAttrs:  common.StringSet{},
		persistentAttrs: common.StringSet{},
		isService:       isService,
		//compositiveMethodComponentIndices: map[string][]int{},
	}

	registeredEntityTypes[typeName] = entityTypeDesc

	entityPtrType := reflect.PtrTo(entityType)
	numMethods := entityPtrType.NumMethod()
	for i := 0; i < numMethods; i++ {
		method := entityPtrType.Method(i)
		rpcDescs.visit(method)
	}

	gwlog.Infof(">>> RegisterEntity %s => %s <<<", typeName, entityType.Name())
	//// define entity attrs
	entity.DescribeEntityType(entityTypeDesc)

	// now entity type is described
	return entityTypeDesc
}

var entityType = reflect.TypeOf(Entity{})

func isEntityType(t reflect.Type) bool {
	if t == entityType {
		return true
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	entityField, ok := t.FieldByName("Entity")
	return ok && entityField.Type == entityType
}

//var componentType = reflect.TypeOf(Component{})

//func isComponentType(t reflect.Type) bool {
//	//if t == componentType {
//	//	return true
//	//}
//	if t.Kind() != reflect.Struct {
//		return false
//	}
//	componentField, ok := t.FieldByName("Component")
//	return ok && componentField.Type == componentType
//}

type createCause int

const (
	ccCreate createCause = 1 + iota
	ccMigrate
	ccRestore
)

func createEntity(typeName string, space *Space, pos Vector3, entityID common.EntityID, data map[string]interface{}, timerData []byte, client *GameClient, cause createCause) common.EntityID {
	//gwlog.Debugf("createEntity: %s in Space %s", typeName, space)
	entityTypeDesc, ok := registeredEntityTypes[typeName]
	if !ok {
		gwlog.Panicf("unknown entity type: %s", typeName)
	}

	if entityID == "" {
		entityID = common.GenEntityID()
	}

	var entity *Entity
	var entityInstance reflect.Value

	entityInstance = reflect.New(entityTypeDesc.entityType)
	entity = reflect.Indirect(entityInstance).FieldByName("Entity").Addr().Interface().(*Entity)
	entity.init(typeName, entityID, entityInstance)
	entity.Space = nilSpace

	entityManager.put(entity)
	if data != nil {
		if cause == ccCreate {
			entity.loadPersistentData(data)
		} else {
			entity.LoadMigrateData(data)
		}
	} else {
		entity.Save() // save immediately after creation
	}

	if timerData != nil {
		entity.restoreTimers(timerData)
	}

	isPersistent := entity.IsPersistent()
	if isPersistent { // startup the periodical timer for saving e
		entity.setupSaveTimer()
	}

	if cause == ccCreate || cause == ccRestore {
		dispatchercluster.SendNotifyCreateEntity(entityID)
	}

	if client != nil {
		// assign client to the newly created
		if cause == ccCreate {
			entity.SetClient(client)
		} else {
			entity.client = client // assign client quietly if migrate
			entityManager.onEntityGetClient(entity.ID, client.clientid)
		}
	}

	gwlog.Debugf("Entity %s created, cause=%d, client=%s", entity, cause, client)
	entity.I.OnAttrsReady()
	//entity.callCompositiveMethod("OnAttrsReady")

	if cause == ccCreate {
		entity.I.OnCreated()
		//entity.callCompositiveMethod("OnCreated")
	} else if cause == ccMigrate {
		entity.I.OnMigrateIn()
		//entity.callCompositiveMethod("OnMigrateIn")
	} else if cause == ccRestore {
		// restore should be silent
		entity.I.OnRestored()
		//entity.callCompositiveMethod("OnRestored")
	}

	if space != nil {
		space.enter(entity, pos, cause == ccRestore)
	}

	return entityID
}

func loadEntityLocally(typeName string, entityID common.EntityID, space *Space, pos Vector3) {
	// load the data from storage
	storage.Load(typeName, entityID, func(data interface{}, err error) {
		// callback runs in main routine
		if err != nil {
			gwlog.Panicf("load entity %s.%s failed: %s", typeName, entityID, err)
			dispatchercluster.SendNotifyDestroyEntity(entityID) // load entity failed, tell dispatcher
		}

		if space != nil && space.IsDestroyed() {
			// Space might be destroy during the Load process, so cancel the entity creation
			dispatchercluster.SendNotifyDestroyEntity(entityID) // load entity failed, tell dispatcher
			return
		}

		createEntity(typeName, space, pos, entityID, data.(map[string]interface{}), nil, nil, ccCreate)
	})
}

func loadEntityAnywhere(typeName string, entityID common.EntityID) {
	dispatchercluster.SendLoadEntityAnywhere(typeName, entityID)
}

func createEntityAnywhere(typeName string, data map[string]interface{}) common.EntityID {
	entityid := common.GenEntityID()
	dispatchercluster.SendCreateEntityAnywhere(entityid, typeName, data)
	return entityid
}

// CreateEntityLocally creates new entity in the local game
func CreateEntityLocally(typeName string, data map[string]interface{}, client *GameClient) common.EntityID {
	entityTypeDesc, ok := registeredEntityTypes[typeName]
	if !ok {
		gwlog.Panicf("unknown entity type: %s", typeName)
	} else if entityTypeDesc.isService {
		gwlog.Panicf("should not create service entity manually: %s", typeName)
	}
	return createEntity(typeName, nil, Vector3{}, "", data, nil, client, ccCreate)
}

// CreateEntityAnywhere creates new entity in any game
func CreateEntityAnywhere(typeName string) common.EntityID {
	entityTypeDesc, ok := registeredEntityTypes[typeName]
	if !ok {
		gwlog.Panicf("unknown entity type: %s", typeName)
	} else if entityTypeDesc.isService {
		gwlog.Panicf("should not create service entity manually: %s", typeName)
	}
	return createEntityAnywhere(typeName, nil)
}

// OnCreateEntityAnywhere is called when CreateEntityAnywhere chooses this game
func OnCreateEntityAnywhere(entityid common.EntityID, typeName string, data map[string]interface{}) {
	createEntity(typeName, nil, Vector3{}, entityid, data, nil, nil, ccCreate)
}

// LoadEntityLocally loads entity in the local game.
//
// LoadEntityLocally has no effect if entity already exists on any game
func LoadEntityLocally(typeName string, entityID common.EntityID) {
	loadEntityLocally(typeName, entityID, nil, Vector3{})
}

// LoadEntityAnywhere loads entity in the any game
//
// LoadEntityAnywhere has no effect if entity already exists on any game
func LoadEntityAnywhere(typeName string, entityID common.EntityID) {
	loadEntityAnywhere(typeName, entityID)
}

// OnClientDisconnected is called by engine when client is disconnected
func OnClientDisconnected(clientid common.ClientID) {
	entityManager.onClientDisconnected(clientid) // pop the owner eid
}

// OnDeclareService is called by engine when service is declared
func OnDeclareService(serviceName string, entityid common.EntityID) {
	entityManager.onDeclareService(serviceName, entityid)
}

// OnUndeclareService is called by engine when service is undeclared
func OnUndeclareService(serviceName string, entityid common.EntityID) {
	entityManager.onUndeclareService(serviceName, entityid)
}

// GetServiceProviders returns all service providers
func GetServiceProviders(serviceName string) EntityIDSet {
	return entityManager.registeredServices[serviceName]
}

func Call(id common.EntityID, method string, args []interface{}) {
	if consts.OPTIMIZE_LOCAL_ENTITY_CALL {
		e := entityManager.get(id)
		if e != nil { // this entity is local, just call entity directly
			e.Post(func() {
				e.onCallFromLocal(method, args)
			})
		} else {
			callRemote(id, method, args)
		}
	} else {
		callRemote(id, method, args)
	}
}

func CallService(serviceName string, method string, args []interface{}) {
	serviceEid := entityManager.chooseServiceProvider(serviceName)
	Call(serviceEid, method, args)
}

func CallNilSpaces(method string, args []interface{}, gameid uint16) {
	if consts.OPTIMIZE_LOCAL_ENTITY_CALL {
		dispatchercluster.SendCallNilSpaces(gameid, method, args)
		nilSpace.onCallFromLocal(method, args)
	} else {
		dispatchercluster.SendCallNilSpaces(0, method, args)
	}
}

func OnCallNilSpaces(method string, args [][]byte) {
	nilSpace.onCallFromRemote(method, args, "")
}

func callRemote(id common.EntityID, method string, args []interface{}) {
	dispatchercluster.SelectByEntityID(id).SendCallEntityMethod(id, method, args)
}

var lastWarnedOnCallMethod = ""

// OnCall is called by engine when method call reaches in the game
func OnCall(id common.EntityID, method string, args [][]byte, clientID common.ClientID) {
	e := entityManager.get(id)
	if e == nil {
		// entity not found, may destroyed before call
		if method != lastWarnedOnCallMethod {
			gwlog.Warnf("OnCall: entity %s is not found while calling %s", id, method)
			lastWarnedOnCallMethod = method
		}

		return
	}

	e.onCallFromRemote(method, args, clientID)
}

// OnSyncPositionYawFromClient is called by engine to sync entity infos from client
func OnSyncPositionYawFromClient(eid common.EntityID, x, y, z Coord, yaw Yaw) {
	e := entityManager.get(eid)
	if e == nil {
		// entity not found, may destroyed before call
		//gwlog.Errorf("OnSyncPositionYawFromClient: entity %s is not found", eid)
		return
	}

	e.syncPositionYawFromClient(x, y, z, yaw)
}

// GetEntity returns the entity with specified ID
func GetEntity(id common.EntityID) *Entity {
	return entityManager.get(id)
}

// OnGameTerminating is called when game is terminating
func OnGameTerminating() {
	for _, e := range entityManager.entities {
		e.Destroy()
	}
}

var allGamesConnected bool

// OnAllGamesConnected is called when all games are connected to dispatcher cluster
func OnAllGamesConnected() {

	allGamesConnected = true
	gwlog.Infof("all games connected, nil space = %s", nilSpace)
	if nilSpace != nil {
		nilSpace.I.OnGameReady()
	}
}

// OnGateDisconnected is called when gate is down
func OnGateDisconnected(gateid uint16) {
	gwlog.Warnf("Gate %d disconnected", gateid)
	entityManager.onGateDisconnected(gateid)
}

// SaveAllEntities saves all entities
func SaveAllEntities() {
	for _, e := range entityManager.entities {
		e.Save()
	}
}

// Called by engine when server is freezing

// FreezeData is the data structure for storing entity freeze data
type FreezeData struct {
	Entities map[common.EntityID]*entityFreezeData
	Services map[string][]common.EntityID
}

// Freeze freezes the entity system and returns all freeze data
func Freeze(gameid uint16) (*FreezeData, error) {
	freeze := FreezeData{}

	entityFreezeInfos := map[common.EntityID]*entityFreezeData{}
	foundNilSpace := false
	for _, e := range entityManager.entities {

		err := gwutils.CatchPanic(func() {
			e.I.OnFreeze()
			//e.callCompositiveMethod("OnFreeze")
		})
		if err != nil {
			// OnFreeze failed
			return nil, errors.Errorf("OnFreeze paniced: %v", err)
		}

		entityFreezeInfos[e.ID] = e.GetFreezeData()
		if e.IsSpaceEntity() {
			if e.ToSpace().IsNil() {
				if foundNilSpace {
					return nil, errors.Errorf("found duplicate nil space")
				}
				foundNilSpace = true
			}
		}
	}

	if !foundNilSpace { // there should be exactly one nil space!
		return nil, errors.Errorf("nil space not found")
	}

	freeze.Entities = entityFreezeInfos
	registeredServices := make(map[string][]common.EntityID, len(entityManager.registeredServices))
	for serviceName, eids := range entityManager.registeredServices {
		registeredServices[serviceName] = eids.ToList()
	}
	freeze.Services = registeredServices

	return &freeze, nil
}

// RestoreFreezedEntities restore entity system from freeze data
func RestoreFreezedEntities(freeze *FreezeData) (err error) {
	defer func() {
		_err := recover()
		if _err != nil {
			err = errors.Wrap(_err.(error), "panic during restore")
		}

	}()

	restoreEntities := func(filter func(typeName string, spaceKind int64) bool) {
		for eid, info := range freeze.Entities {
			typeName := info.Type
			var spaceKind int64
			if typeName == _SPACE_ENTITY_TYPE {
				attrs := info.Attrs
				spaceKind = typeconv.Int(attrs[_SPACE_KIND_ATTR_KEY])
			}

			if filter(typeName, spaceKind) {
				var space *Space
				if typeName != _SPACE_ENTITY_TYPE {
					space = spaceManager.getSpace(info.SpaceID)
				}

				var client *GameClient
				if info.Client != nil {
					client = MakeGameClient(info.Client.ClientID, info.Client.GateID)
				}
				createEntity(typeName, space, info.Pos, eid, info.Attrs, info.TimerData, client, ccRestore)
				gwlog.Debugf("Restored %s<%s> in space %s", typeName, eid, space)

				if info.ESR != nil { // entity was entering space before freeze, so restore entering space
					post.Post(func() {
						entity := GetEntity(eid)
						if entity != nil {
							entity.EnterSpace(info.ESR.SpaceID, info.ESR.EnterPos)
						}
					})
				}
			}
		}
	}
	// step 1: restore the nil space
	restoreEntities(func(typeName string, spaceKind int64) bool {
		return typeName == _SPACE_ENTITY_TYPE && spaceKind == 0
	})

	// step 2: restore all other spaces
	restoreEntities(func(typeName string, spaceKind int64) bool {
		return typeName == _SPACE_ENTITY_TYPE && spaceKind != 0
	})

	// step  3: restore all other spaces
	restoreEntities(func(typeName string, spaceKind int64) bool {
		return typeName != _SPACE_ENTITY_TYPE
	})

	for serviceName, _eids := range freeze.Services {
		eids := EntityIDSet{}
		for _, eid := range _eids {
			eids.Add(eid)
		}
		entityManager.registeredServices[serviceName] = eids
	}

	return nil
}

// Entities gets all entities
//
// Never modify the return value !
func Entities() EntityMap {
	return entityManager.entities
}
