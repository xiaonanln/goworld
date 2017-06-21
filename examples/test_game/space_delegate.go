package main

import (
	. "github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type SpaceDelegate struct {
	DefaultSpaceDelegate // override from default space delegate
}

func (delegate *SpaceDelegate) OnSpaceCreated(space *Space) {
	delegate.DefaultSpaceDelegate.OnSpaceCreated(space)

	//avatarIds := goworld.ListEntityIDs("Avatar")[:1]
	//for _, avatarID := range avatarIds {
	//	gwlog.Info("Loading avatar %s", avatarID)
	//	space.LoadEntity("Avatar", avatarID)
	//}
	//N := 20 - len(avatarIds)
	//for i := 0; i < N; i++ {
	//	space.CreateEntity("Avatar")
	//}

	// notify the SpaceService that it's ok
	space.CallService("SpaceService", "NotifySpaceLoaded", space.Kind, space.ID)

	M := 10
	for i := 0; i < M; i++ {
		space.CreateEntity("Monster")
	}
}

func (delegate *SpaceDelegate) OnEntityEnterSpace(space *Space, entity *Entity) {
	if entity.TypeName == "Avatar" {
		space.clearCheckDestroyTimer()
	}
}

func (delegate *SpaceDelegate) OnEntityLeaveSpace(space *Space, entity *Entity) {
	if entity.TypeName == "Avatar" {
		delegate.onAvatarLeaveSpace(space, entity)
	}
}

func (delegate *SpaceDelegate) onAvatarLeaveSpace(space *Space, entity *Entity) {
	gwlog.Info("Avatar %s leave space %s, left avatar count %d", entity, space, space.CountEntities("Avatar"))
	if space.CountEntities("Avatar") == 0 {
		// no avatar left, start destroying space
		space.CallService("SpaceService", "RequestDestroy", space.ID)
	}
}
