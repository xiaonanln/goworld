package main

import (
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type SpaceService struct {
	entity.Entity
}

func (s *SpaceService) OnInit() {
}

func (s *SpaceService) OnCreated() {
	gwlog.Info("Registering SpaceService ...")
	s.DeclareService("SpaceService")
}

func (s *SpaceService) IsPersistent() bool {
	return false
}
