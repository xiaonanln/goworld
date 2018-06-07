package service

import (
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/srvdis"
)

const (
	checkServicesInterval = time.Second * 5
)

var (
	serviceTypes = map[string]*entity.EntityTypeDesc{}
)

func RegisterService(typeName string, entityPtr entity.IEntity) {
	td := entity.RegisterEntity(typeName, entityPtr)
	serviceTypes[typeName] = td

}

func init() {
	timer.AddTimer(checkServicesInterval, checkServices)
}

func checkServices() {
	gwlog.Infof("checking services ...")
	aliveServices := common.StringSet{}
	srvdis.VisitServicesByTypePrefix("service/", func(srvtype, srvid string, info srvdis.ServiceRegisterInfo) {
		gwlog.Infof("service: found service %s.%s with register info %s", srvtype, srvid, info)
		serviceType := getServiceType(srvtype)
		aliveServices.Add(serviceType)
	})

	for serviceType := range serviceTypes {
		if !aliveServices.Contains(serviceType) {
			gwlog.Warnf("service: %s not found, try to create ...", serviceType)
		}
	}
}

func getServiceType(srvtype string) string {
	return srvtype[len("service/"):]
}
