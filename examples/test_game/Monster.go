package main

import "github.com/xiaonanln/goworld/entity"

type Monster struct {
	entity.Entity
}

func (e *Monster) OnCreated() {
	e.Entity.OnCreated()
	//e.AddTimer(time.Second, func() {
	//	gwlog.Info("%s.Neighbors = %v", e, e.Neighbors())
	//	for _other := range e.Neighbors() {
	//		if _other.TypeName != "Monster" {
	//			continue
	//		}
	//
	//		other := _other.I.(*Monster)
	//		gwlog.Info("%s is a neighbor of %s", other, e)
	//	}
	//})
}
