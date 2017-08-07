package entity

import (
	"fmt"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
)

const (
	_SPACE_ENTITY_TYPE   = "__space__"
	_SPACE_KIND_ATTR_KEY = "_K"

	_DEFAULT_AOI_DISTANCE = 100
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
	aoiCalc  AOICalculator
}

func init() {

}

func (space *Space) String() string {
	if space.Kind != 0 {
		return fmt.Sprintf("Space<%d|%s>", space.Kind, space.ID)
	}
	return "Space<nil>"
}

// OnInit initialize Space entity
func (space *Space) OnInit() {
	space.entities = EntitySet{}
	space.I = space.Entity.I.(ISpace)
	space.aoiCalc = newXZListAOICalculator()
	gwutils.RunPanicless(space.I.OnSpaceInit)
}

// OnSpaceInit is called when Space is initializing
//
// Custom space type can override to provide custom logic
func (space *Space) OnSpaceInit() {

}

// OnCreated is called when Space entity is created
func (space *Space) OnCreated() {
	//dispatcher_client.GetDispatcherClientForSend().SendNotifyCreateEntity(space.ID)
	space.onSpaceCreated()
	if space.IsNil() {
		return
	}

	if consts.DEBUG_SPACES {
		gwlog.Debug("%s.OnCreated", space)
	}
	gwutils.RunPanicless(space.I.OnSpaceCreated)
}

// OnRestored is called when space entity is restored
func (space *Space) OnRestored() {
	space.onSpaceCreated()
}

func (space *Space) onSpaceCreated() {
	space.Kind = space.GetInt(_SPACE_KIND_ATTR_KEY)
	spaceManager.putSpace(space)

	if space.Kind == 0 {
		if nilSpace != nil {
			gwlog.Panicf("duplicate nil space: %s && %s", nilSpace, space)
		}
		nilSpace = space
		nilSpace.Space = nilSpace
		gwlog.Info("Created nil space: %s", nilSpace)
		return
	}
}

// OnSpaceCreated is called when space is created
//
// Custom space type can override to provide custom logic
func (space *Space) OnSpaceCreated() {
	if consts.DEBUG_SPACES {
		gwlog.Debug("Space %s created", space)
	}
}

// OnDestroy is called when Space entity is destroyed
func (space *Space) OnDestroy() {
	gwutils.RunPanicless(space.I.OnSpaceDestroy)
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
		gwlog.Debug("Space %s created", space)
	}
}

// IsNil checks if the space is the nil space
func (space *Space) IsNil() bool {
	return space.Kind == 0
}

// CreateEntity creates a new local entity in this space
func (space *Space) CreateEntity(typeName string, pos Position) {
	createEntity(typeName, space, pos, "", nil, nil, nil, ccCreate)
}

// LoadEntity loads a entity of specified entityID to the space
//
// If the entity already exists on server, this call has no effect
func (space *Space) LoadEntity(typeName string, entityID common.EntityID, pos Position) {
	loadEntityLocally(typeName, entityID, space, pos)
}

func (space *Space) enter(entity *Entity, pos Position, isRestore bool) {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s.enter <<< %s, avatar count=%d, monster count=%d", space, entity, space.CountEntities("Avatar"), space.CountEntities("Monster"))
	}

	if entity.Space != nilSpace {
		gwlog.Panicf("%s.enter(%s): current Space is not nil", space, entity)
	}

	if space.IsNil() { // enter nil space does nothing
		return
	}

	entity.Space = space
	space.entities.Add(entity)

	space.aoiCalc.Enter(&entity.aoi, pos)
	entity.syncInfoFlag |= sifSyncOwnClient | sifSyncNeighborClients

	if !isRestore {
		entity.client.sendCreateEntity(&space.Entity, false) // create Space entity before every other entities

		enter, _ := space.aoiCalc.Adjust(&entity.aoi)
		// gwlog.Info("Entity %s entering at pos %v: %v: enter %d neighbors", entity, pos, entity.GetPosition(), len(enter))

		for _, naoi := range enter {
			neighbor := naoi.getEntity()
			entity.interest(neighbor)
			neighbor.interest(entity)
		}

		gwutils.RunPanicless(func() {
			space.I.OnEntityEnterSpace(entity)
			entity.I.OnEnterSpace()
		})
	} else {
		enter, _ := space.aoiCalc.Adjust(&entity.aoi)
		for _, naoi := range enter {
			neighbor := naoi.getEntity()
			entity.aoi.interest(neighbor)
			neighbor.aoi.interest(entity)
		}
	}

	//space.verifyAOICorrectness(entity)
}

func (space *Space) leave(entity *Entity) {
	if entity.Space != space {
		gwlog.Panicf("%s.leave(%s): entity is not in this Space", space, entity)
	}

	if space.IsNil() { // leave from nil space do nothing
		return
	}

	for neighbor := range entity.aoi.neighbors {
		entity.uninterest(neighbor)
		neighbor.uninterest(entity)
	}
	space.aoiCalc.Leave(&entity.aoi)
	entity.client.sendDestroyEntity(&space.Entity)
	// remove from Space entities
	space.entities.Del(entity)
	entity.Space = nilSpace

	gwutils.RunPanicless(func() {
		space.I.OnEntityLeaveSpace(entity)
	})

	entity.I.OnLeaveSpace(space)
}

func (space *Space) move(entity *Entity, newPos Position) {
	space.aoiCalc.Move(&entity.aoi, newPos)
	enter, leave := space.aoiCalc.Adjust(&entity.aoi)

	for _, naoi := range leave {
		neighbor := naoi.getEntity()
		entity.uninterest(neighbor)
		neighbor.uninterest(entity)
	}

	for _, naoi := range enter {
		neighbor := naoi.getEntity()
		entity.interest(neighbor)
		neighbor.interest(entity)
	}

	//space.verifyAOICorrectness(entity)
	//opmon.Finish(time.Millisecond * 10)
}

//func (space *Space) verifyAOICorrectness(entity *Entity) {
//	if space.IsNil() {
//		return
//	}
//
//	for e := range space.entities {
//		if e.aoi.markVal != 0 {
//			gwlog.Fatalf("%s: wrong AOI mark val = %d", e.aoi.markVal)
//		}
//
//		if e == entity {
//			continue
//		}
//
//		isNeighbor := e.aoi.pos.X >= entity.aoi.pos.X-_DEFAULT_AOI_DISTANCE && e.aoi.pos.X <= entity.aoi.pos.X+_DEFAULT_AOI_DISTANCE && e.aoi.pos.Z >= entity.aoi.pos.Z-_DEFAULT_AOI_DISTANCE && e.aoi.pos.Z <= entity.aoi.pos.Z+_DEFAULT_AOI_DISTANCE
//		if entity.aoi.neighbors.Contains(e) && !isNeighbor {
//			gwlog.Fatalf("space %s: %s: wrong neighbor: %s, pos=%v, %v", space, entity, e, entity.GetPosition(), e.GetPosition())
//		} else if !entity.aoi.neighbors.Contains(e) && isNeighbor {
//			gwlog.Fatalf("space %s: %s: wrong not neighbor: %s: pos=%v, %v", space, entity, e, entity.GetPosition(), e.GetPosition())
//		}
//	}
//}

// OnEntityEnterSpace is called when entity enters space
//
// Custom space type can override this function
func (space *Space) OnEntityEnterSpace(entity *Entity) {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s ENTER SPACE %s", entity, space)
	}
}

// OnEntityLeaveSpace is called when entity leaves space
//
// Custom space type can override this function
func (space *Space) OnEntityLeaveSpace(entity *Entity) {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s LEAVE SPACE %s", entity, space)
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

// AOI Management
func (space *Space) addToAOI(entity *Entity) {

}
