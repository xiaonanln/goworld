package main

import (
	"time"

	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
)

type OnlineService struct {
	entity.Entity
}

func (s *OnlineService) OnCreated() {
	s.AddCallback(time.Second*3, func() {
		gwlog.Info("Registering OnlineService ...")
		s.RegisterService("OnlineService")
	})
}
