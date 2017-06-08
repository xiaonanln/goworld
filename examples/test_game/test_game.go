package main

import (
	"time"

	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/components/game"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type Monster struct {
	entity.Entity
}

func init() {

}

type gameDelegate struct {
	game.GameDelegate
}

func main() {
	goworld.SetSpaceDelegate(&SpaceDelegate{})
	goworld.RegisterEntity("Monster", &Monster{})
	goworld.Run(&gameDelegate{})
}

func (game gameDelegate) OnReady() {
	game.GameDelegate.OnReady()
	// create the space
	goworld.CreateSpace()
	//eid1 := goworld.createEntity("Monster")
	//eid2 := goworld.createEntity("Monster")
}

type SpaceDelegate struct {
	entity.DefaultSpaceDelegate // override from default space delegate
}

func (delegate *SpaceDelegate) OnSpaceCreated(space *entity.Space) {
	delegate.DefaultSpaceDelegate.OnSpaceCreated(space)

	N := 3
	for i := 0; i < N; i++ {
		space.CreateEntity("Monster")
	}
}

func (e *Monster) OnCreated() {
	e.Entity.OnCreated()
	gwlog.Info("Creating callback ...")
	e.AddTimer(time.Second, func() {
		gwlog.Info("%s.Neighbors = %v", e, e.Neighbors())
		for _other := range e.Neighbors() {
			if _other.TypeName != "Monster" {
				continue
			}

			other := _other.I.(*Monster)
			gwlog.Info("%s is a neighbor of %s", other, e)
		}
	})
}
