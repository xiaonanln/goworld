package main

import (
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/consts"
	. "github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

const (
	SPACE_DESTROY_CHECK_INTERVAL = time.Minute * 5
)

type MySpace struct {
	Space

	destroyCheckTimer *timer.Timer
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
		space.onAvatarEnterSpace(entity)
	}
}

func (space *MySpace) onAvatarEnterSpace(entity *Entity) {
	space.clearDestroyCheckTimer()
}

func (space *MySpace) OnEntityLeaveSpace(entity *Entity) {
	if entity.TypeName == "Avatar" {
		space.onAvatarLeaveSpace(entity)
	}
}

func (space *MySpace) onAvatarLeaveSpace(entity *Entity) {
	if consts.DEBUG_SPACES {
		gwlog.Info("Avatar %s leave space %s, left avatar count %d", entity, space, space.CountEntities("Avatar"))
	}
	if space.CountEntities("Avatar") == 0 {
		// no avatar left, start destroying space
		space.setDestroyCheckTimer()
	}
}

func (space *MySpace) setDestroyCheckTimer() {
	if space.destroyCheckTimer != nil {
		return
	}

	space.destroyCheckTimer = space.AddTimer(SPACE_DESTROY_CHECK_INTERVAL, func() {
		avatarCount := space.CountEntities("Avatar")
		if avatarCount != 0 {
			gwlog.Panicf("Avatar count should be 0, but is %d", avatarCount)
		}

		space.CallService("SpaceService", "RequestDestroy", space.Kind, space.ID)
	})
}

func (space *MySpace) clearDestroyCheckTimer() {
	if space.destroyCheckTimer == nil {
		return
	}

	space.CancelTimer(space.destroyCheckTimer)
	space.destroyCheckTimer = nil
}

func (space *MySpace) ConfirmRequestDestroy_Server(ok bool) {
	if ok {
		if space.CountEntities("Avatar") != 0 {
			gwlog.Panicf("%s ConfirmRequestDestroy: avatar count is %d", space, space.CountEntities("Avatar"))
		}
		space.Destroy()
	}
}
