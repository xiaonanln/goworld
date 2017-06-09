package main

import (
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type OnlineService struct {
	entity.Entity
}

func (s *OnlineService) OnCreated() {
	gwlog.Info("Registering OnlineService ...")
	s.DeclareService("OnlineService")
}

func (s *OnlineService) CheckIn_Server() {

}
