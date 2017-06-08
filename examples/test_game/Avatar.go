package main

import "github.com/xiaonanln/goworld/entity"

type Avatar struct {
	entity.Entity
}

func (e *Avatar) OnCreated() {
	e.Entity.OnCreated()
}
