package entity

import (
	"fmt"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/gwutils"
)

const (
	SPACE_ENTITY_TYPE   = "__space__"
	SPACE_KIND_ATTR_KEY = "_K"
)

var (
	nilSpace *Space
)

type Space struct {
	Entity

	entities EntitySet
	Kind     int
	I        ISpace
}

func init() {

}

func (space *Space) String() string {
	if space.Kind != 0 {
		return fmt.Sprintf("Space<%d|%s>", space.Kind, space.ID)
	} else {
		return "Space<nil>"
	}
}

func (space *Space) OnInit() {
	space.entities = EntitySet{}
	space.I = space.Entity.I.(ISpace)
	gwutils.RunPanicless(space.I.OnSpaceInit)
}

func (space *Space) OnSpaceInit() {

}

func (space *Space) OnCreated() {
	space.Kind = space.GetInt(SPACE_KIND_ATTR_KEY)
	spaceManager.putSpace(space)

	if space.Kind == 0 {
		if nilSpace != nil {
			gwlog.Panicf("duplicate nil space: %s && %s", nilSpace, space)
		}
		nilSpace = space
		return
	}

	//dispatcher_client.GetDispatcherClientForSend().SendNotifyCreateEntity(space.ID)
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s.OnCreated", space)
	}
	gwutils.RunPanicless(space.I.OnSpaceCreated)
}

func (space *Space) OnSpaceCreated() {
	if consts.DEBUG_SPACES {
		gwlog.Debug("Space %s created", space)
	}
}

func (space *Space) OnDestroy() {
	gwutils.RunPanicless(space.I.OnSpaceDestroy)
	// destroy all entities
	for e := range space.entities {
		e.Destroy()
	}

	spaceManager.delSpace(space.ID)
}

func (space *Space) OnSpaceDestroy() {
	if consts.DEBUG_SPACES {
		gwlog.Debug("Space %s created", space)
	}
}

func (space *Space) IsNil() bool {
	return space.Kind == 0
}

func (space *Space) CreateEntity(typeName string) {
	createEntity(typeName, space, "", nil, nil, false)
}

func (space *Space) LoadEntity(typeName string, entityID common.EntityID) {
	loadEntityLocally(typeName, entityID, space)
}

func (space *Space) enter(entity *Entity) {
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
	entity.interest(&space.Entity) // interest the Space entity before every other entities
	for other := range space.entities {
		entity.interest(other)
		other.interest(entity)
	}
	space.entities.Add(entity)
	gwutils.RunPanicless(func() {
		space.I.OnEntityEnterSpace(entity)
	})

	entity.I.OnEnterSpace()
}

func (space *Space) leave(entity *Entity) {
	if entity.Space != space {
		gwlog.Panicf("%s.leave(%s): e is not in this Space", space, entity)
	}

	if space.IsNil() { // leave from nil space do nothing
		return
	}

	entity.Space = nilSpace
	// remove from Space entities
	space.entities.Del(entity)
	for other := range space.entities {
		entity.uninterest(other)
		other.uninterest(entity)
	}
	entity.uninterest(&space.Entity)

	gwutils.RunPanicless(func() {
		space.I.OnEntityLeaveSpace(entity)
	})

	entity.I.OnLeaveSpace(space)
}

func (space *Space) OnEntityEnterSpace(entity *Entity) {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s ENTER SPACE %s", entity, space)
	}
}

func (space *Space) OnEntityLeaveSpace(entity *Entity) {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s LEAVE SPACE %s", entity, space)
	}
}

func (space *Space) CountEntities(typeName string) int {
	count := 0
	for e, _ := range space.entities {
		if e.TypeName == typeName {
			count += 1
		}
	}
	return count
}

func (space *Space) GetEntityCount() int {
	return len(space.entities)
}
