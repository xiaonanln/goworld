package service

import (
	"time"

	"fmt"

	"strings"

	"strconv"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/srvdis"
	"github.com/xiaonanln/goworld/engine/storage"
)

const (
	checkServicesInterval  = time.Second * 5
	serviceSrvdisPrefix    = "service/"
	serviceSrvdisPrefixLen = len(serviceSrvdisPrefix)
)

var (
	registeredServices = map[string]*entity.EntityTypeDesc{}
	gameid             uint16
)

func RegisterService(typeName string, entityPtr entity.IEntity) {
	td := entity.RegisterEntity(typeName, entityPtr)
	registeredServices[typeName] = td
}

func Startup(gameid_ uint16) {
	gameid = gameid_
	timer.AddTimer(checkServicesInterval, checkServices)
}

func checkServices() {
	gwlog.Infof("checking services ...")
	registeredServices := common.StringSet{}
	needLocalServiceEntities := common.StringSet{}

	srvdis.TraverseByPrefix(serviceSrvdisPrefix, func(srvid string, srvinfo string) {
		servicePath := strings.Split(srvid[serviceSrvdisPrefixLen:], "/")
		gwlog.Infof("service: found service %v = %+v", servicePath, srvinfo)

		if len(servicePath) == 1 {
			serviceName := servicePath[0]
			targetGameID, err := strconv.Atoi(srvinfo[4:])
			if err != nil {
				gwlog.Panic(errors.Wrap(err, "parse targetGameID failed"))
			}
			// XxxService = gameX
			registeredServices.Add(serviceName)

			if int(gameid) == targetGameID {
				needLocalServiceEntities.Add(serviceName)
				serviceEntities := entity.GetEntitiesByType(serviceName)
				gwlog.Infof("service: %s should be created on game%d, local entities: %+v", serviceName, targetGameID, serviceEntities)
			}
		}
	})

	for serviceName := range needLocalServiceEntities {
		serviceEntities := entity.GetEntitiesByType(serviceName)
		if len(serviceEntities) == 0 {
			createServiceEntity(serviceName)
		}
	}

	for serviceName := range registeredServices {
		if !registeredServices.Contains(serviceName) {
			gwlog.Warnf("service: %s not found, registering srvdis ...", serviceName)
			srvdis.Register(getSrvID(serviceName), fmt.Sprintf("game%d", gameid))
		}
	}
}
func createServiceEntity(serviceName string) {
	storage.ListEntityIDs(serviceName, func(eids []common.EntityID, err error) {
		gwlog.Infof("Found saved %s ids: %v", serviceName, eids)

		if len(eids) == 0 {
			goworld.CreateEntityLocally(serviceName)
		} else {
			// already exists
			serviceID := eids[0]
			goworld.LoadEntityAnywhere(serviceName, serviceID)
		}
	})
}

func getSrvID(serviceName string) string {
	return serviceSrvdisPrefix + serviceName
}
