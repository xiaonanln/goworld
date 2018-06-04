package service

import (
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

var (
	serviceTypes = map[string]*entity.EntityTypeDesc{}
)

func RegisterService(typeName string, entityPtr entity.IEntity) {
	td := entity.RegisterEntity(typeName, entityPtr)
	serviceTypes[typeName] = td

}

func init() {
	timer.AddTimer(time.Second, tickPerSecond)
	//go gwutils.RepeatUntilPanicless(func() {
	//	serviceMgrRoutine()
	//})
}

func tickPerSecond() {
	gwlog.Infof("discover service ...")

}
