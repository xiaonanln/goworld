package service

import (
	"fmt"
	"time"

	"strings"

	"strconv"

	"math/rand"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwvar"
	"github.com/xiaonanln/goworld/engine/kvreg"
)

const (
	checkServicesInterval      = time.Second * 60
	serviceKvregPrefix         = "Service/"
	serviceKvregPrefixLen      = len(serviceKvregPrefix)
	checkServicesLaterDelayMax = time.Millisecond * 500
	serviceNameShardIndexSep   = "#" // must not be "/"
	maxServiceShardCount       = 8192
)

type serviceId string

func getServiceId(serviceName string, shardIndex int) serviceId {
	return serviceId(fmt.Sprintf("%s#%d", serviceName, shardIndex))
}

func splitServiceId(serviceId serviceId) (serviceName string, shardIndex int) {
	var err error

	serviceNameIndex := strings.Split(string(serviceId), serviceNameShardIndexSep)
	if len(serviceNameIndex) != 2 {
		gwlog.Panicf("invalid service id: %#v (should be ServiceName#ShardIndex)", serviceId)
	}

	serviceName = serviceNameIndex[0]
	shardIndex, err = strconv.Atoi(serviceNameIndex[1])
	if err != nil {
		gwlog.Panicf("invalid service id: %#v (invalid shard index)", serviceId)
	}

	if shardIndex < 0 || shardIndex > maxServiceShardCount {
		gwlog.Panicf("invalid service id: %#v (shard index is out of range)", serviceId)
	}

	return
}

var (
	registeredServices = map[string]int{}
	gameid             uint16
	serviceMap         = map[string][]common.EntityID{} // ServiceName -> []Entity ID
	checkTimer         *timer.Timer
)

func RegisterService(typeName string, entityPtr entity.IEntity, shardCount int) {
	if shardCount <= 0 || shardCount > maxServiceShardCount {
		gwlog.Panicf("RegisterService: %s is using invalid shard count: %d, should be in range [%d ~ %d]", typeName, shardCount, 1, maxServiceShardCount)
	}

	if strings.Contains(typeName, serviceNameShardIndexSep) {
		gwlog.Panicf("RegisterService: invalid service name: %s (should not contains %#v)", typeName, serviceNameShardIndexSep)
	}

	entity.RegisterEntity(typeName, entityPtr, true)
	registeredServices[typeName] = shardCount
}

func Setup(gameid_ uint16) {
	gameid = gameid_
	kvreg.AddPostCallback(checkServicesLater)
}

func OnDeploymentReady() {
	timer.AddTimer(checkServicesInterval, checkServicesLater)
	checkServicesLater()
}

type serviceInfo struct {
	Registered bool
	EntityID   common.EntityID
}

func checkServicesLater() {
	if checkTimer != nil {
		checkTimer.Cancel()
		checkTimer = nil
	}

	checkTimer = timer.AddCallback(time.Duration(rand.Int63n(int64(checkServicesLaterDelayMax))), func() {
		checkTimer = nil
		checkServices()
	})
}

// checkServices is executed per `checkServicesInterval`
func checkServices() {
	if !gwvar.IsDeploymentReady.Value() {
		// deployment is not ready
		return
	}
	gwlog.Infof("service: checking services ...")
	dispRegisteredServices := map[serviceId]*serviceInfo{}     // all services that are registered on dispatchers
	localRegServiceIds := map[serviceId]struct{}{}             //service ids that are registered on this game server
	localRegServiceEntities := map[string]common.EntityIDSet{} // local service entities that is registered, group by ServiceName
	newServiceMap := make(map[string][]common.EntityID, len(registeredServices))

	// ServiceId == ServiceName#ShardIndex
	getServiceInfo := func(serviceId serviceId) *serviceInfo {
		info := dispRegisteredServices[serviceId]
		if info == nil {
			info = &serviceInfo{}
			dispRegisteredServices[serviceId] = info
		}
		return info
	}

	kvreg.TraverseByPrefix(serviceKvregPrefix, func(key string, val string) {
		servicePath := strings.Split(key[serviceKvregPrefixLen:], "/")
		//gwlog.Infof("service: found service %v = %+v", servicePath, val)

		if len(servicePath) == 1 {
			// ServiceName#ShardIndex = gameX
			serviceId := serviceId(servicePath[0])
			regGameId, err := strconv.Atoi(val[4:])
			if err != nil {
				gwlog.Panic(errors.Wrap(err, "parse gameid failed"))
			}
			getServiceInfo(serviceId).Registered = true

			// this service entity should be created on local game server
			if int(gameid) == regGameId {
				localRegServiceIds[serviceId] = struct{}{}
			}
		} else if len(servicePath) == 2 {
			// ServiceName#ShardIndex/EntityID = Xxxx
			serviceId := serviceId(servicePath[0])
			fieldName := servicePath[1]
			switch fieldName {
			case "EntityID":
				getServiceInfo(serviceId).EntityID = common.EntityID(val)
			default:
				gwlog.Errorf("unknown kvreg info: %s = %s", key, val)
			}
		} else {
			gwlog.Errorf("unknown kvreg key: %s", servicePath)
		}
	})

	// generate new service map from registered service IDs
	for serviceId, info := range dispRegisteredServices {
		if !info.Registered || info.EntityID.IsNil() {
			continue
		}

		serviceName, shardIndex := splitServiceId(serviceId)
		regShardCount := registeredServices[serviceName]

		if shardIndex >= regShardCount {
			gwlog.Errorf("invalid service id: %#v (shard index is out of range)", serviceId)
			continue
		}

		serviceEids := newServiceMap[serviceName]
		if serviceEids == nil {
			serviceEids = make([]common.EntityID, regShardCount)
			newServiceMap[serviceName] = serviceEids
		}

		serviceEids[shardIndex] = info.EntityID
	}
	// replace with new service map
	serviceMap = newServiceMap

	// find all service entities that should be created on local game, group by service name
	for serviceId := range localRegServiceIds {
		serviceInfo := getServiceInfo(serviceId)
		serviceName, _ := splitServiceId(serviceId)
		if !serviceInfo.EntityID.IsNil() {
			if localRegServiceEntities[serviceName] == nil {
				localRegServiceEntities[serviceName] = common.EntityIDSet{}
			}

			localRegServiceEntities[serviceName].Add(serviceInfo.EntityID)
		}
	}

	// destroy all service entities that is on this game, but is not registered successfully
	for serviceName := range registeredServices {
		serviceEntities := entity.GetEntitiesByType(serviceName)
		for eid, entity := range serviceEntities {
			if !localRegServiceEntities[serviceName].Contains(eid) {
				// this service entity is created locally, but not registered
				// might be caused by registration delay (very low chance)
				entity.Destroy()
			}
		}
	}

	// create all service entities that should be created on this game
	for serviceId := range localRegServiceIds {
		serviceInfo := getServiceInfo(serviceId)

		if serviceInfo.EntityID.IsNil() || entity.GetEntity(serviceInfo.EntityID) == nil {
			// service entity not created locally yet
			createServiceEntity(serviceId)
		}
	}

	// register all service ids that are not registered to dispatcher yet
	for serviceName, shardCount := range registeredServices {
		for shardIndex := 0; shardIndex < shardCount; shardIndex++ {
			serviceId := getServiceId(serviceName, shardIndex)
			serviceInfo := getServiceInfo(serviceId)

			if serviceInfo.Registered {
				continue
			}

			gwlog.Warnf("service: %s not found, registering kvreg ...", serviceId)

			// delay for a random time so that each game might register services randomly
			randomDelay := time.Millisecond * time.Duration(rand.Intn(1000))
			timer.AddCallback(randomDelay, func() {
				kvreg.Register(getServiceRegKey(serviceId), fmt.Sprintf("game%d", gameid), false)
			})
		}
	}
}

func createServiceEntity(serviceId serviceId) {
	serviceName, shardIndex := splitServiceId(serviceId)
	_ = shardIndex

	desc := entity.GetEntityTypeDesc(serviceName)
	if desc == nil {
		gwlog.Panicf("create service entity locally failed: service %s is not registered", serviceName)
	}

	e := entity.CreateEntityLocally(serviceName, nil)
	kvreg.Register(getServiceRegKey(serviceId)+"/EntityID", string(e.ID), true)
	gwlog.Infof("Created service entity: %s: %s", serviceName, e)
}

func getServiceRegKey(serviceId serviceId) string {
	return serviceKvregPrefix + string(serviceId)
}

func CallServiceAny(serviceName string, method string, args []interface{}) {
	serviceEids := serviceMap[serviceName]
	if len(serviceEids) == 0 {
		gwlog.Errorf("CallServiceAny %s.%s: no service entity found!", serviceName, method)
		return
	}

	eid := serviceEids[rand.Intn(len(serviceEids))]
	if eid.IsNil() {
		gwlog.Errorf("CallServiceAny %s.%s: service entity is nil!", serviceName, method)
		return
	}

	entity.Call(eid, method, args)
}

func CallServiceAll(serviceName string, method string, args []interface{}) {
	serviceEids := serviceMap[serviceName]
	if len(serviceEids) == 0 {
		gwlog.Errorf("CallServiceAll %s.%s: no service entity found!", serviceName, method)
		return
	}

	for shardIndex, eid := range serviceEids {
		if eid.IsNil() {
			gwlog.Errorf("CallServiceAll %s.%s: service entity %d is nil!", serviceName, method, shardIndex)
			continue
		}

		// TODO: optimize calls to multiple entities
		entity.Call(eid, method, args)
	}
}

func CallServiceShardIndex(serviceName string, shardIndex int, method string, args []interface{}) {
	serviceEids := serviceMap[serviceName]
	if shardIndex < 0 || shardIndex >= len(serviceEids) {
		gwlog.Errorf("CallServiceShardIndex %s.%s: found %d service entities, but shard index is %d!", serviceName, method, len(serviceEids), shardIndex)
		return
	}

	eid := serviceEids[shardIndex]
	if eid.IsNil() {
		gwlog.Errorf("CallServiceShardIndex %s.%s: service entity %d is nil!", serviceName, method, shardIndex)
		return
	}

	entity.Call(eid, method, args)
}

func CallServiceShardKey(serviceName string, shardKey string, method string, args []interface{}) {
	serviceEids := serviceMap[serviceName]

	if len(serviceEids) <= 0 {
		gwlog.Errorf("CallServiceShardKey %s.%s: no service entities", serviceName, method)
		return
	}

	shardIndex := shardByKey(shardKey, len(serviceEids))
	eid := serviceEids[shardIndex]
	if eid.IsNil() {
		gwlog.Errorf("CallServiceShardKey %s.%s: service entity %d (shard key %+v) is nil!", serviceName, method, shardIndex, shardKey)
		return
	}

	entity.Call(eid, method, args)
}

func shardByKey(key string, shardCount int) int {
	return int(common.HashString(key)) % shardCount
}

func GetServiceEntityID(serviceName string, shardIndex int) common.EntityID {
	eids := serviceMap[serviceName]
	if shardIndex >= 0 && shardIndex < len(eids) {
		return eids[shardIndex]
	} else {
		return ""
	}
}

func GetServiceShardCount(serviceName string) int {
	return registeredServices[serviceName]
}

func CheckServiceEntitiesReady(serviceName string) bool {
	shardCount := registeredServices[serviceName]
	gwlog.Warnf("CheckServiceEntitiesReady %s: shard=%d, eids=%+v", serviceName, shardCount, serviceMap[serviceName])
	if shardCount <= 0 {
		return false
	}

	eids := serviceMap[serviceName]
	if len(eids) != shardCount {
		return false
	}

	for _, eid := range eids {
		if eid.IsNil() {
			return false
		}
	}

	return true
}
