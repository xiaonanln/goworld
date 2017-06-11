package main

import "github.com/xiaonanln/goworld/entity"

type Monster struct {
	entity.Entity
}

func (e *Monster) OnCreated() {
	e.Entity.OnCreated()
	//a.AddTimer(time.Second, func() {
	//	gwlog.Info("%s.Neighbors = %v", s, s.Neighbors())
	//	for _other := range a.Neighbors() {
	//		if _other.TypeName != "Monster" {
	//			continue
	//		}
	//
	//		other := _other.I.(*Monster)
	//		gwlog.Info("%s is s neighbor of %s", other, s)
	//	}
	//})
}
