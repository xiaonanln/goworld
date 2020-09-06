package entity

import (
	"reflect"

	"strings"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/dispatchercluster"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/storage"
	"github.com/xiaonanln/typeconv"
)

var (
	registeredEntityTypes = map[string]*EntityTypeDesc{}
	entityManager         = newEntityManager()
)

// EntityTypeDesc is the entity type description for registering entity types
type EntityTypeDesc struct {
	isService       bool
	IsPersistent    bool
	useAOI          bool
	aoiDistance     Coord
	entityType      reflect.Type
	rpcDescs        rpcDescMap
	allClientAttrs  common.StringSet
	clientAttrs     common.StringSet
	persistentAttrs common.StringSet
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
	if desc.isService {
		gwlog.Panicf("Service entity must NOT be persistent: %s", desc.entityType.Name())
	}

	desc.IsPersistent = persistent
	return desc
}

func (desc *EntityTypeDesc) SetUseAOI(useAOI bool, aoiDistance Coord) *EntityTypeDesc {
	if aoiDistance < 0 {
		gwlog.Panicf("aoi distance < 0")
	}

	desc.useAOI = useAOI
	desc.aoiDistance = aoiDistance
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
			if !desc.IsPersistent {
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
	entities       EntityMap
	entitiesByType map[string]EntityMap
}

func newEntityManager() *_EntityManager {
	return &_EntityManager{
		entities:       EntityMap{},
		entitiesByType: map[string]EntityMap{},
	}
}

func (em *_EntityManager) put(entity *Entity) {
	em.entities.Add(entity)
	etype := entity.TypeName
	eid := entity.ID
	if entities, ok := em.entitiesByType[etype]; ok {
		entities.Add(entity)
	} else {
		em.entitiesByType[etype] = EntityMap{eid: entity}
	}
}

func (em *_EntityManager) del(e *Entity) {
	eid := e.ID
	em.entities.Del(eid)
	if entities, ok := em.entitiesByType[e.TypeName]; ok {
		entities.Del(eid)
	}
}

func (em *_EntityManager) get(id common.EntityID) *Entity {
	return em.entities.Get(id)
}

func (em *_EntityManager) traverseByType(etype string, cb func(e *Entity)) {
	entities := em.entitiesByType[etype]
	for _, e := range entities {
		cb(e)
	}
}

func (em *_EntityManager) onGateDisconnected(gateid uint16) {
	for _, entity := range em.entities {
		client := entity.client
		if client != nil && client.gateid == gateid {
			entity.notifyClientDisconnected()
		}
	}
}

// RegisterEntity registers custom entity type and define entity behaviors
func RegisterEntity(typeName string, entity IEntity, isService bool) *EntityTypeDesc {
	if _, ok := registeredEntityTypes[typeName]; ok {
		gwlog.Fatalf("RegisterEntity: Entity type %s already registered", typeName)
	}

	entityVal := reflect.ValueOf(entity)
	entityType := entityVal.Type()

	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	// register the string of e
	rpcDescs := rpcDescMap{}
	entityTypeDesc := &EntityTypeDesc{
		isService:       isService,
		IsPersistent:    false,
		useAOI:          false,
		entityType:      entityType,
		rpcDescs:        rpcDescs,
		clientAttrs:     common.StringSet{},
		allClientAttrs:  common.StringSet{},
		persistentAttrs: common.StringSet{},
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
	//// define entity Attrs
	entity.DescribeEntityType(entityTypeDesc)
	return entityTypeDesc
}

func GetEntityTypeDesc(typeName string) *EntityTypeDesc {
	return registeredEntityTypes[typeName]
}

var entityType = reflect.TypeOf(Entity{})

func createEntity(typeName string, space *Space, pos Vector3, entityID common.EntityID, data map[string]interface{}) *Entity {
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
		entity.loadPersistentData(data)
	} else {
		entity.Save() // save immediately after creation
	}

	isPersistent := entity.IsPersistent()
	if isPersistent { // startup the periodical timer for saving e
		entity.setupSaveTimer()
	}

	dispatchercluster.SendNotifyCreateEntity(entityID)

	gwlog.Debugf("Entity %s created.", entity)
	gwutils.RunPanicless(func() {
		entity.I.OnAttrsReady()
		entity.I.OnCreated()
	})

	if space != nil {
		space.enter(entity, pos, false)
	}

	return entity
}

func restoreEntity(entityID common.EntityID, mdata *entityMigrateData, isRestore bool) {
	typeName := mdata.Type
	entityTypeDesc, ok := registeredEntityTypes[typeName]
	if !ok {
		gwlog.Panicf("unknown entity type: %s", typeName)
	}

	var entity *Entity
	var entityInstance reflect.Value

	entityInstance = reflect.New(entityTypeDesc.entityType)
	entity = reflect.Indirect(entityInstance).FieldByName("Entity").Addr().Interface().(*Entity)
	entity.init(typeName, entityID, entityInstance)
	entity.Space = nilSpace
	entity.Position = mdata.Pos
	entity.yaw = mdata.Yaw

	entityManager.put(entity)
	entity.loadMigrateData(mdata.Attrs)

	timerData := mdata.TimerData
	if timerData != nil {
		entity.restoreTimers(timerData)
	}

	isPersistent := entity.IsPersistent()
	if isPersistent { // startup the periodical timer for saving e
		entity.setupSaveTimer()
	}

	entity.syncInfoFlag = mdata.SyncInfoFlag
	entity.syncingFromClient = mdata.SyncingFromClient

	if mdata.Client != nil {
		client := MakeGameClient(mdata.Client.ClientID, mdata.Client.GateID)
		// assign Client to the newly created
		entity.assignClient(client) // assign Client quietly
	}

	gwlog.Debugf("Entity %s created, Client=%s", entity, entity.client)
	gwutils.RunPanicless(func() {
		entity.I.OnAttrsReady()
	})

	if !isRestore {
		gwutils.RunPanicless(func() {
			entity.I.OnMigrateIn()
		})
	}
	space := spaceManager.getSpace(mdata.SpaceID)
	if space != nil {
		space.enter(entity, mdata.Pos, isRestore)
	}

	if isRestore {
		gwutils.RunPanicless(func() {
			entity.I.OnRestored()
		})
	}
}

func loadEntityLocally(typeName string, entityID common.EntityID, space *Space, pos Vector3) {
	// load the data from storage
	storage.Load(typeName, entityID, func(_data interface{}, err error) {
		// callback runs in main routine
		if err != nil {
			dispatchercluster.SendNotifyDestroyEntity(entityID) // load entity failed, tell dispatcher
			gwlog.Panicf("load entity %s.%s failed: %s", typeName, entityID, err)
		}

		ex := entityManager.get(entityID) // existing entity
		if ex != nil {
			// should not happen because dispatcher won't allow, but just in case
			gwlog.Panicf("load entity %s.%s failed: %s already exists", typeName, entityID, ex)
		}

		if space != nil && space.IsDestroyed() {
			// Space might be destroy during the Load process, so cancel the entity creation
			space = nil // if space is destroyed before creation, just use nil space
		}

		data := _data.(map[string]interface{})
		// need to remove NOT persistent fields from data
		entityTypeDesc := registeredEntityTypes[typeName]
		removeFields := []string{}
		for k := range data {
			if !entityTypeDesc.persistentAttrs.Contains(k) {
				removeFields = append(removeFields, k)
			}
		}
		for _, f := range removeFields {
			delete(data, f)
		}
		createEntity(typeName, space, pos, entityID, data)
	})
}

func createEntitySomewhere(gameid uint16, typeName string, data map[string]interface{}) common.EntityID {
	entityid := common.GenEntityID()
	dispatchercluster.SendCreateEntitySomewhere(gameid, entityid, typeName, data)
	return entityid
}

// CreateEntityLocally creates new entity in the local game
func CreateEntityLocally(typeName string, data map[string]interface{}) *Entity {
	return createEntity(typeName, nil, Vector3{}, "", data)
}

// CreateEntityLocallyWithEntityID creates new entity in the local game with specified entity ID
func CreateEntityLocallyWithID(typeName string, data map[string]interface{}, id common.EntityID) *Entity {
	return createEntity(typeName, nil, Vector3{}, id, data)
}

// CreateEntitySomewhere creates new entity in any game
func CreateEntitySomewhere(gameid uint16, typeName string) common.EntityID {
	return createEntitySomewhere(gameid, typeName, nil)
}

// OnCreateEntitySomewhere is called when CreateEntitySomewhere chooses this game
func OnCreateEntitySomewhere(entityid common.EntityID, typeName string, data map[string]interface{}) {
	createEntity(typeName, nil, Vector3{}, entityid, data)
}

// OnLoadEntitySomewhere loads entity in the local game.
func OnLoadEntitySomewhere(typeName string, entityID common.EntityID) {
	loadEntityLocally(typeName, entityID, nil, Vector3{})
}

// LoadEntityAnywhere loads entity in the any game
//
// LoadEntityAnywhere has no effect if entity already exists on any game
func LoadEntityAnywhere(typeName string, entityID common.EntityID) {
	dispatchercluster.SendLoadEntityAnywhere(typeName, entityID)
}

func LoadEntityOnGame(typeName string, entityID common.EntityID, gameid uint16) {
	dispatchercluster.SendLoadEntityOnGame(typeName, entityID, gameid)
}

// OnClientDisconnected is called by engine when Client is disconnected
func OnClientDisconnected(ownerID common.EntityID, clientid common.ClientID) {
	owner := entityManager.get(ownerID)
	if owner != nil {
		if owner.client != nil && owner.client.clientid == clientid {
			owner.notifyClientDisconnected()
		} else {
			gwlog.Warnf("client %s is disconnected, but owner entity %s has client %s", clientid, owner, owner.client)
		}
	} else {
		gwlog.Warnf("owner entity %s not found for client %s, might already be destroyed", ownerID, clientid)
	}
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

// OnSyncPositionYawFromClient is called by engine to sync entity infos from Client
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

func GetEntitiesByType(etype string) EntityMap {
	return entityManager.entitiesByType[etype]
}

// TraverseEntityByType traverses entities of the specified type
func TraverseEntityByType(etype string, cb func(e *Entity)) {
	entityManager.traverseByType(etype, cb)
}

// OnGameTerminating is called when game is terminating
func OnGameTerminating() {
	for _, e := range entityManager.entities {
		e.Destroy()
	}
}

var gameIsReady bool

// OnGameReady is called when all games are connected to dispatcher cluster
func OnGameReady() {
	if gameIsReady {
		gwlog.Warnf("all games connected, but not for the first time")
		//gwlog.Warnf("registered services: %+v", entityManager.registeredServices)
		return
	}

	gameIsReady = true
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
	Entities map[common.EntityID]*entityMigrateData
}

// Freeze freezes the entity system and returns all freeze data
func Freeze(gameid uint16) (*FreezeData, error) {
	freeze := FreezeData{}

	entityFreezeInfos := map[common.EntityID]*entityMigrateData{}
	foundNilSpace := false
	for _, e := range entityManager.entities {

		err := gwutils.CatchPanic(func() {
			e.I.OnFreeze()
		})
		if err != nil {
			// OnFreeze failed
			return nil, errors.Errorf("OnFreeze paniced: %v", err)
		}

		entityFreezeInfos[e.ID] = e.getFreezeData()
		if e.IsSpaceEntity() {
			if e.AsSpace().IsNil() {
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

	clients := map[common.EntityID]*GameClient{} // save all the clients when restoring entities
	restoreEntities := func(filter func(typeName string, spaceKind int64) bool) {
		for eid, info := range freeze.Entities {
			typeName := info.Type
			var spaceKind int64
			if typeName == _SPACE_ENTITY_TYPE {
				spaceKind = typeconv.Int(info.Attrs[_SPACE_KIND_ATTR_KEY])
			}

			if filter(typeName, spaceKind) {
				var space *Space
				if typeName != _SPACE_ENTITY_TYPE {
					space = spaceManager.getSpace(info.SpaceID)
				}

				var client *GameClient
				if info.Client != nil {
					client = MakeGameClient(info.Client.ClientID, info.Client.GateID)
					clients[eid] = client // save the Client to the map
					info.Client = nil
				}
				restoreEntity(eid, info, true)
				gwlog.Debugf("Restored %s<%s> in space %s", typeName, eid, space)
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

	// restore clients to all entities
	for eid, client := range clients {
		e := entityManager.get(eid)
		if e != nil {
			e.assignClient(client) // assign Client quietly if migrate
		} else {
			gwlog.Errorf("entity %s restore failed? can not set Client %s", eid, client)
		}
	}

	return nil
}

// Entities gets all entities
//
// Never modify the return value !
func Entities() EntityMap {
	return entityManager.entities
}
