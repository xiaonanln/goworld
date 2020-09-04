package entity

import (
	"fmt"

	"github.com/xiaonanln/go-aoi"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
)

const (
	_SPACE_ENTITY_TYPE    = "__space__"
	_SPACE_KIND_ATTR_KEY  = "_K"
	_SPACE_ENABLE_AOI_KEY = "_EnableAOI"
)

var (
	nilSpace *Space
)

// Space is the entity type of spaces
//
// Spaces are also entities but with space management logics
type Space struct {
	Entity

	entities EntitySet
	Kind     int
	I        ISpace

	aoiMgr aoi.AOIManager
}

func (space *Space) String() string {
	if space == nil {
		return "nil"
	}

	if space.Kind != 0 {
		return fmt.Sprintf("Space<%d|%s>", space.Kind, space.ID)
	} else {
		return fmt.Sprintf("NilSpace<%s>", space.ID)
	}
}

func (space *Space) DescribeEntityType(desc *EntityTypeDesc) {
	desc.DefineAttr(_SPACE_KIND_ATTR_KEY, "AllClients")
}

func (space *Space) GetSpaceRange() (minX, minY, maxX, maxY Coord) {
	return -1000, -1000, 1000, 1000
}

func (space *Space) GetTowerRange() (minX, minY, maxX, maxY Coord) {
	return -1000, -1000, 1000, 1000
}

// OnInit initialize Space entity
func (space *Space) OnInit() {
	space.entities = EntitySet{}
	space.I = space.Entity.I.(ISpace)

	space.I.OnSpaceInit()
}

// OnSpaceInit is a compositive method for initializing space fields
func (space *Space) OnSpaceInit() {

}

// OnCreated is called when Space entity is created
func (space *Space) OnCreated() {
	//dispatcher_client.GetDispatcherClientForSend().SendNotifyCreateEntity(space.ID)
	space.onSpaceCreated()
	if space.IsNil() {
		gwlog.Infof("nil space is created: %s, all games connected: %v", space, gameIsReady)
		if gameIsReady {
			space.I.OnGameReady()
		}
		return
	}

	if consts.DEBUG_SPACES {
		gwlog.Debugf("%s.OnCreated", space)
	}
	space.I.OnSpaceCreated()
}

func (space *Space) EnableAOI(defaultAOIDistance Coord) {
	if defaultAOIDistance <= 0 {
		gwlog.Panicf("defaultAOIDistance < 0")
	}

	if space.aoiMgr != nil {
		gwlog.Panicf("%s.EnableAOI: AOI already enabled", space)
	}

	if len(space.entities) > 0 {
		gwlog.Panicf("%s is already using AOI", space)
	}

	space.Attrs.SetFloat(_SPACE_ENABLE_AOI_KEY, float64(defaultAOIDistance))
	space.aoiMgr = aoi.NewXZListAOIManager(aoi.Coord(defaultAOIDistance))
}

// OnRestored is called when space entity is restored
func (space *Space) OnRestored() {
	space.onSpaceCreated()
	//gwlog.Debugf("space %s restored: atts=%+v", space, space.Attrs)
	aoidist := space.GetFloat(_SPACE_ENABLE_AOI_KEY)
	if aoidist > 0 {
		space.EnableAOI(Coord(aoidist))
	}
}

func (space *Space) onSpaceCreated() {
	space.Kind = int(space.GetInt(_SPACE_KIND_ATTR_KEY))
	spaceManager.putSpace(space)

	if space.Kind == 0 {
		if nilSpace != nil {
			gwlog.Panicf("duplicate nil space: %s && %s", nilSpace, space)
		}
		nilSpace = space
		nilSpace.Space = nilSpace
		gwlog.Infof("Created nil space: %s", nilSpace)
		return
	}
}

// OnSpaceCreated is called when space is created
//
// Custom space type can override to provide custom logic
func (space *Space) OnSpaceCreated() {
	if consts.DEBUG_SPACES {
		gwlog.Debugf("Space %s created", space)
	}
}

// OnDestroy is called when Space entity is destroyed
func (space *Space) OnDestroy() {
	space.I.OnSpaceDestroy()
	// destroy all entities
	for e := range space.entities {
		e.Destroy()
	}

	spaceManager.delSpace(space.ID)
}

// OnSpaceDestroy is called when space is destroying
//
// Custom space type can override to provide custom logic
func (space *Space) OnSpaceDestroy() {
	if consts.DEBUG_SPACES {
		gwlog.Debugf("Space %s created", space)
	}
}

// IsNil checks if the space is the nil space
func (space *Space) IsNil() bool {
	return space.Kind == 0
}

// CreateEntity creates a new local entity in this space
func (space *Space) CreateEntity(typeName string, pos Vector3) {
	createEntity(typeName, space, pos, "", nil)
}

// LoadEntity loads a entity of specified entityID to the space
//
// If the entity already exists on server, this call has no effect
func (space *Space) LoadEntity(typeName string, entityID common.EntityID, pos Vector3) {
	loadEntityLocally(typeName, entityID, space, pos)
}

func (space *Space) enter(entity *Entity, pos Vector3, isRestore bool) {
	if consts.DEBUG_SPACES {
		gwlog.Debugf("%s.enter <<< %s, avatar count=%d, monster count=%d", space, entity, space.CountEntities("Avatar"), space.CountEntities("Monster"))
	}

	if entity.Space != nilSpace {
		gwlog.Panicf("%s.enter(%s): current space is not nil, but %s", space, entity, entity.Space)
	}

	if space.IsNil() { // enter nil space does nothing
		return
	}

	entity.Space = space
	space.entities.Add(entity)
	entity.Position = pos

	entity.syncInfoFlag |= sifSyncOwnClient | sifSyncNeighborClients

	if !isRestore {
		entity.client.sendCreateEntity(&space.Entity, false) // create Space entity before every other entities

		if space.aoiMgr != nil && entity.IsUseAOI() {
			space.aoiMgr.Enter(&entity.aoi, aoi.Coord(pos.X), aoi.Coord(pos.Z))
		}

		gwutils.RunPanicless(func() {
			space.I.OnEntityEnterSpace(entity)
			entity.I.OnEnterSpace()
		})
	} else {
		// restoring ...
		if space.aoiMgr != nil && entity.IsUseAOI() {
			space.aoiMgr.Enter(&entity.aoi, aoi.Coord(pos.X), aoi.Coord(pos.Z))
		}

	}
	//space.verifyAOICorrectness(entity)
}

func (space *Space) leave(entity *Entity) {
	if entity.Space != space {
		gwlog.Panicf("%s.leave(%s): entity is not in this Space", space, entity)
	}

	if space.IsNil() {
		// leaving nil space does nothing
		return
	}

	// remove from Space entities
	space.entities.Del(entity)
	entity.Space = nilSpace

	if space.aoiMgr != nil && entity.IsUseAOI() {
		space.aoiMgr.Leave(&entity.aoi)
	}

	entity.client.sendDestroyEntity(&space.Entity)
	gwutils.RunPanicless(func() {
		space.I.OnEntityLeaveSpace(entity)
		entity.I.OnLeaveSpace(space)
	})
}

func (space *Space) move(entity *Entity, newPos Vector3) {
	if space.aoiMgr == nil {
		return
	}

	entity.Position = newPos
	space.aoiMgr.Moved(&entity.aoi, aoi.Coord(newPos.X), aoi.Coord(newPos.Z))
	gwlog.Debugf("%s: %s move to %v", space, entity, newPos)
}

// OnEntityEnterSpace is called when entity enters space
//
// Custom space type can override this function
func (space *Space) OnEntityEnterSpace(entity *Entity) {
	if consts.DEBUG_SPACES {
		gwlog.Debugf("%s ENTER SPACE %s", entity, space)
	}
}

// OnEntityLeaveSpace is called when entity leaves space
//
// Custom space type can override this function
func (space *Space) OnEntityLeaveSpace(entity *Entity) {
	if consts.DEBUG_SPACES {
		gwlog.Debugf("%s LEAVE SPACE %s", entity, space)
	}
}

// CountEntities returns the number of entities of specified type in space
func (space *Space) CountEntities(typeName string) int {
	count := 0
	for e := range space.entities {
		if e.TypeName == typeName {
			count += 1
		}
	}
	return count
}

// GetEntityCount returns the total count of entities in space
func (space *Space) GetEntityCount() int {
	return len(space.entities)
}

// ForEachEntity visits all entities in space and call function f with each entity
func (space *Space) ForEachEntity(f func(e *Entity)) {
	for e := range space.entities {
		f(e)
	}
}

// GetEntity returns the entity in space with specified ID, nil otherwise
func (space *Space) GetEntity(entityID common.EntityID) *Entity {
	entity := GetEntity(entityID)
	if entity == nil {
		return nil
	}

	if space.entities.Contains(entity) {
		return entity
	} else {
		return nil
	}
}

// aoi Management
func (space *Space) addToAOI(entity *Entity) {

}

// OnGameReady is called when the game server is ready on NilSpace only
func (space *Space) OnGameReady() {
	gwlog.Warnf("Game server is ready. Override function %T.OnGameReady to write your own game logic!", space.I)
}
