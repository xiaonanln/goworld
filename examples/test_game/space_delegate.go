package main

import (
	. "github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type MySpace struct {
	Space
}

func (space *MySpace) OnSpaceCreated() {
	// notify the SpaceService that it's ok
	space.CallService("SpaceService", "NotifySpaceLoaded", space.Kind, space.ID)

	M := 10
	for i := 0; i < M; i++ {
		space.CreateEntity("Monster")
	}
}

func (space *MySpace) OnEntityEnterSpace(entity *Entity) {
	if entity.TypeName == "Avatar" {
	}
}

func (space *MySpace) OnEntityLeaveSpace(entity *Entity) {
	if entity.TypeName == "Avatar" {
		space.onAvatarLeaveSpace(entity)
	}
}

func (space *MySpace) onAvatarLeaveSpace(entity *Entity) {
	gwlog.Info("Avatar %s leave space %s, left avatar count %d", entity, space, space.CountEntities("Avatar"))
	if space.CountEntities("Avatar") == 0 {
		// no avatar left, start destroying space
		space.CallService("SpaceService", "RequestDestroy", space.ID)
	}
}
