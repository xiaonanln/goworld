package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type SpaceDelegate struct {
	entity.DefaultSpaceDelegate // override from default space delegate
}

func (delegate *SpaceDelegate) OnSpaceCreated(space *entity.Space) {
	delegate.DefaultSpaceDelegate.OnSpaceCreated(space)

	avatarIds := goworld.ListEntityIDs("Avatar")[:1]
	for _, avatarID := range avatarIds {
		gwlog.Info("Loading avatar %s", avatarID)
		space.LoadEntity("Avatar", avatarID)
	}
	N := 20 - len(avatarIds)
	for i := 0; i < N; i++ {
		space.CreateEntity("Avatar")
	}

	M := 10
	for i := 0; i < M; i++ {
		space.CreateEntity("Monster")
	}

}
