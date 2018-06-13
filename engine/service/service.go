package service

import (
	"time"

	"fmt"

	"strings"

	"strconv"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/srvdis"
)

const (
	checkServicesInterval  = time.Second * 5
	serviceSrvdisPrefix    = "Service/"
	serviceSrvdisPrefixLen = len(serviceSrvdisPrefix)
)

var (
	registeredServices = common.StringSet{}
	gameid             uint16
	serviceMap         = map[string]common.EntityID{} // ServiceName -> Entity ID
)

func RegisterService(typeName string, entityPtr entity.IEntity) {
	entity.RegisterEntity(typeName, entityPtr)
	registeredServices.Add(typeName)
}

func Startup(gameid_ uint16) {
	gameid = gameid_
	timer.AddTimer(checkServicesInterval, checkServices)
}

type serviceInfo struct {
	Registered bool
	EntityID   common.EntityID
}

func checkServices() {
	gwlog.Infof("checking services ...")
	dispRegisteredServices := map[string]*serviceInfo{} // all services that are registered on dispatchers
	needLocalServiceEntities := common.StringSet{}
	newServiceMap := make(map[string]common.EntityID, len(registeredServices))

	getServiceInfo := func(serviceName string) *serviceInfo {
		info := dispRegisteredServices[serviceName]
		if info == nil {
			info = &serviceInfo{}
			dispRegisteredServices[serviceName] = info
		}
		return info
	}

	srvdis.TraverseByPrefix(serviceSrvdisPrefix, func(srvid string, srvinfo string) {
		servicePath := strings.Split(srvid[serviceSrvdisPrefixLen:], "/")
		//gwlog.Infof("service: found service %v = %+v", servicePath, srvinfo)

		if len(servicePath) == 1 {
			// ServiceName = gameX
			serviceName := servicePath[0]
			targetGameID, err := strconv.Atoi(srvinfo[4:])
			if err != nil {
				gwlog.Panic(errors.Wrap(err, "parse targetGameID failed"))
			}
			// XxxService = gameX
			getServiceInfo(serviceName).Registered = true

			if int(gameid) == targetGameID {
				needLocalServiceEntities.Add(serviceName)
			}
		} else if len(servicePath) == 2 {
			// ServiceName/EntityID = Xxxx
			serviceName := servicePath[0]
			fieldName := servicePath[1]
			switch fieldName {
			case "EntityID":
				getServiceInfo(serviceName).EntityID = common.EntityID(srvinfo)
			default:
				gwlog.Warnf("unknown srvdis info: %s = %s", srvid, srvinfo)
			}
		} else {
			gwlog.Panic(servicePath)
		}
	})

	for serviceName, info := range dispRegisteredServices {
		if info.Registered && !info.EntityID.IsNil() {
			newServiceMap[serviceName] = info.EntityID
		}
	}
	serviceMap = newServiceMap

	for serviceName := range needLocalServiceEntities {
		serviceEntities := entity.GetEntitiesByType(serviceName)
		if len(serviceEntities) == 0 {
			createServiceEntity(serviceName)
		} else if len(serviceEntities) == 1 {
			// make sure the current service entity is the
			localEid := serviceEntities.Keys()[0]
			//gwlog.Infof("service %s: found service entity: %s, service info: %+v", serviceName, serviceEntities.Values()[0], getServiceInfo(serviceName))
			if localEid != getServiceInfo(serviceName).EntityID {
				// might happen if dispatchers recover from crash
				gwlog.Warnf("service %s: local entity is %s, but has %s on dispatchers", serviceName, localEid, getServiceInfo(serviceName).EntityID)
				srvdis.Register(getSrvID(serviceName)+"/EntityID", string(localEid), true)
			}
		} else {
			// multiple service entities ? should never happen! so just destroy all invalid service entities
			correctEid := getServiceInfo(serviceName).EntityID
			for _, e := range serviceEntities {
				if e.ID != correctEid {
					e.Destroy()
				}
			}
		}
	}

	for serviceName := range registeredServices {
		if !getServiceInfo(serviceName).Registered {
			gwlog.Warnf("service: %s not found, registering srvdis ...", serviceName)
			srvdis.Register(getSrvID(serviceName), fmt.Sprintf("game%d", gameid), false)
		}
	}
}
func createServiceEntity(serviceName string) {
	eid := entity.CreateEntityLocally(serviceName, nil, nil)
	gwlog.Infof("Created service entity: %s: %s", serviceName, eid)
	srvdis.Register(getSrvID(serviceName)+"/EntityID", string(eid), true)
}

func getSrvID(serviceName string) string {
	return serviceSrvdisPrefix + serviceName
}

func CallService(serviceName string, method string, args []interface{}) {
	serviceEid := serviceMap[serviceName]
	if serviceEid.IsNil() {
		gwlog.Errorf("CallService %s.%s: service entity is not created yet!", serviceName, method)
		return
	}

	entity.Call(serviceEid, method, args)
}

func GetServiceEntityID(serviceName string) common.EntityID {
	return serviceMap[serviceName]
}
