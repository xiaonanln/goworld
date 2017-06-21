package entity

import (
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
)

var (
	spaceDelegate ISpaceDelegate = &DefaultSpaceDelegate{}
)

func SetSpaceDelegate(delegate ISpaceDelegate) {
	spaceDelegate = delegate
}

// Space delegate interface
type ISpaceDelegate interface {
	OnSpaceCreated(space *Space)
	OnEntityEnterSpace(space *Space, entity *Entity)
	OnEntityLeaveSpace(space *Space, entity *Entity)
}

// The default space delegate
type DefaultSpaceDelegate struct {
}

func (delegate *DefaultSpaceDelegate) OnSpaceCreated(space *Space) {
	if consts.DEBUG_SPACES {
		gwlog.Debug("SPACE CREATED: %s", space)
	}
}

func (delegate *DefaultSpaceDelegate) OnEntityEnterSpace(space *Space, entity *Entity) {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s ENTER SPACE %s", entity, space)
	}
}

func (delegate *DefaultSpaceDelegate) OnEntityLeaveSpace(space *Space, entity *Entity) {
	if consts.DEBUG_SPACES {
		gwlog.Debug("%s LEAVE SPACE %s", entity, space)
	}
}
