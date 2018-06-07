package binutil

import (
	"context"

	"time"

	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/srvdis"
)

const (
	checkServicesInterval = time.Second * 3
)

func StartupCheckServices(ctx context.Context, componentType string, componentId string, compInfo srvdis.ServiceRegisterInfo) {
	go gwutils.RepeatUntilPanicless(ctx, func() {
		checkServicesRoutine(componentType, componentId, compInfo)
	})
}

func checkServicesRoutine(componentType, componentId string, compInfo srvdis.ServiceRegisterInfo) {
	for {
		checkComponentService(componentType, componentId, compInfo)
		time.Sleep(checkServicesInterval)
	}
}

func checkComponentService(componentType, componentId string, info srvdis.ServiceRegisterInfo) {
	componentServiceFound := false
	isComponentServiceMine := true
	srvdis.VisitServicesByType(componentType, func(srvid string, info srvdis.ServiceRegisterInfo) {
		if srvid == componentId {
			componentServiceFound = true
			isComponentServiceMine = info.IsMyLease
		}
	})

	gwlog.Debugf("checkComponentService: %s %s found %v isMine %v", componentType, componentId, componentServiceFound, isComponentServiceMine)
	if !componentServiceFound || !isComponentServiceMine {
		srvdis.Register(componentType, componentId, info, false)
	}
}
