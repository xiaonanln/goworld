package entity

import (
	"fmt"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
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
}

func init() {
	RegisterEntity(SPACE_ENTITY_TYPE, &Space{})
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

	gwlog.Debug("%s.OnCreated", space)
	space.Post(func() {
		spaceDelegate.OnSpaceCreated(space)
	})
}

func (space *Space) OnDestroy() {
	spaceManager.delSpace(space.ID)
}

func (space *Space) IsNil() bool {
	return space.Kind == 0
}

func (space *Space) CreateEntity(typeName string) {
	createEntity(typeName, space, "", nil, nil)
}

func (space *Space) LoadEntity(typeName string, entityID common.EntityID) {
	loadEntityLocally(typeName, entityID, space)
}

func (space *Space) enter(entity *Entity) {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s.enter <<< %s", space, entity)
	}

	if entity.Space != nilSpace {
		gwlog.Panicf("%s.enter(%s): current Space is not nil", space, entity)
	}

	if space.IsNil() { // enter nil space does nothing
		return
	}

	entity.Space = space
	for other := range space.entities {
		entity.interest(other)
		other.interest(entity)
	}
	space.entities.Add(entity)

	entity.I.OnEnterSpace()
}

func (space *Space) leave(entity *Entity) {
	if entity.Space != space {
		gwlog.Panicf("%s.leave(%s): entity is not in this Space", space, entity)
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

	entity.I.OnLeaveSpace(space)
}
