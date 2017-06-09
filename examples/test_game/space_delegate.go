package main

import "github.com/xiaonanln/goworld/entity"

type SpaceDelegate struct {
	entity.DefaultSpaceDelegate // override from default space delegate
}

func (delegate *SpaceDelegate) OnSpaceCreated(space *entity.Space) {
	delegate.DefaultSpaceDelegate.OnSpaceCreated(space)

	N := 3
	for i := 0; i < N; i++ {
		space.CreateEntity("Avatar")
	}

	M := 10
	for i := 0; i < M; i++ {
		space.CreateEntity("Monster")
	}

}
