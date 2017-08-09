package main

import (
	"time"

	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

const (
	_SPACE_DESTROY_CHECK_INTERVAL = time.Minute * 5
)

// MySpace is the custom space type
type MySpace struct {
	entity.Space // Space type should always inherit from entity.Space

	destroyCheckTimer entity.EntityTimerID
}

// OnSpaceCreated is called when the space is created
func (space *MySpace) OnSpaceCreated() {
	// notify the SpaceService that it's ok
	space.CallService("SpaceService", "NotifySpaceLoaded", space.Kind, space.ID)

	M := 10
	for i := 0; i < M; i++ {
		space.CreateEntity("Monster", entity.Position{})
	}
}

// OnEntityEnterSpace is called when entity enters space
func (space *MySpace) OnEntityEnterSpace(entity *entity.Entity) {
	if entity.TypeName == "Avatar" {
		space.onAvatarEnterSpace(entity)
	}
}

func (space *MySpace) onAvatarEnterSpace(entity *entity.Entity) {
	space.clearDestroyCheckTimer()
}

// OnEntityLeaveSpace is called when entity leaves space
func (space *MySpace) OnEntityLeaveSpace(entity *entity.Entity) {
	if entity.TypeName == "Avatar" {
		space.onAvatarLeaveSpace(entity)
	}
}

func (space *MySpace) onAvatarLeaveSpace(entity *entity.Entity) {
	if consts.DEBUG_SPACES {
		gwlog.Infof("Avatar %s leave space %s, left avatar count %d", entity, space, space.CountEntities("Avatar"))
	}
	if space.CountEntities("Avatar") == 0 {
		// no avatar left, start destroying space
		space.setDestroyCheckTimer()
	}
}

func (space *MySpace) setDestroyCheckTimer() {
	if space.destroyCheckTimer != 0 {
		return
	}

	space.destroyCheckTimer = space.AddTimer(_SPACE_DESTROY_CHECK_INTERVAL, "CheckForDestroy")
}

// CheckForDestroy checks if the space should be destroyed
func (space *MySpace) CheckForDestroy() {
	avatarCount := space.CountEntities("Avatar")
	if avatarCount != 0 {
		gwlog.Panicf("Avatar count should be 0, but is %d", avatarCount)
	}

	space.CallService("SpaceService", "RequestDestroy", space.Kind, space.ID)
}

func (space *MySpace) clearDestroyCheckTimer() {
	if space.destroyCheckTimer == 0 {
		return
	}

	space.CancelTimer(space.destroyCheckTimer)
	space.destroyCheckTimer = 0
}

// ConfirmRequestDestroy is called by SpaceService to confirm that the space
func (space *MySpace) ConfirmRequestDestroy(ok bool) {
	if ok {
		if space.CountEntities("Avatar") != 0 {
			gwlog.Panicf("%s ConfirmRequestDestroy: avatar count is %d", space, space.CountEntities("Avatar"))
		}
		space.Destroy()
	}
}
