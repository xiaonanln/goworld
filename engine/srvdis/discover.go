package srvdis

import (
	"encoding/json"

	"strings"

	"sync"

	"sync/atomic"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type serviceTypeMgr struct {
	services map[string]ServiceRegisterInfo
}

func newServiceTypeMgr() *serviceTypeMgr {
	return &serviceTypeMgr{
		services: map[string]ServiceRegisterInfo{},
	}
}

func (mgr *serviceTypeMgr) registerService(srvid string, info ServiceRegisterInfo) {
	mgr.services[srvid] = info
}

func (mgr *serviceTypeMgr) unregisterService(srvid string) {
	delete(mgr.services, srvid)
}

var (
	aliveServicesLock   sync.RWMutex
	aliveServicesByType = map[string]*serviceTypeMgr{}
)

func VisitServicesByType(srvtype string, cb func(srvid string, info ServiceRegisterInfo)) {
	aliveServicesLock.RLock()
	defer aliveServicesLock.RUnlock()

	mgr := aliveServicesByType[srvtype]
	if mgr == nil {
		return
	}

	for srvid, info := range mgr.services {
		cb(srvid, info)
	}
}

func VisitServicesByTypePrefix(srvtypePrefix string, cb func(srvtype, srvid string, info ServiceRegisterInfo)) {
	aliveServicesLock.RLock()
	defer aliveServicesLock.RUnlock()
	for srvtype, mgr := range aliveServicesByType {
		if !strings.HasPrefix(srvtype, srvtypePrefix) {
			continue
		}

		for srvid, info := range mgr.services {
			cb(srvtype, srvid, info)
		}
	}
}

func watchRoutine() {
	kv := clientv3.NewKV(srvdisClient)
	if srvdisNamespace != "" {
		kv = namespace.NewKV(kv, srvdisNamespace)
	}

	rangeResp, err := kv.Get(srvdisCtx, "/srvdis/", clientv3.WithPrefix())
	if err != nil {
		gwlog.Fatal(err)
	}

	for _, kv := range rangeResp.Kvs {
		handlePutServiceRegisterData(serviceDelegate, kv.Key, kv.Value, kv.Lease)
	}

	w := clientv3.NewWatcher(srvdisClient)
	if srvdisNamespace != "" {
		w = namespace.NewWatcher(w, srvdisNamespace)
	}

	ch := w.Watch(srvdisCtx, "/srvdis/", clientv3.WithPrefix(), clientv3.WithRev(rangeResp.Header.Revision+1))
	for resp := range ch {
		for _, event := range resp.Events {
			if event.Type == mvccpb.PUT {
				//gwlog.Infof("watch resp: %v, created=%v, cancelled=%v, events=%q", resp, resp.Created, resp.Canceled, resp.Events[0].Kv.Key)
				handlePutServiceRegisterData(serviceDelegate, event.Kv.Key, event.Kv.Value, event.Kv.Lease)
			} else if event.Type == mvccpb.DELETE {
				gwlog.Warnf("DELETE Kv %+v Lease %d PrevKv %+v", event.Kv, event.Kv.Lease, event.PrevKv)
				handleDeleteServiceRegisterData(serviceDelegate, event.Kv.Key)
			}
		}
	}
}

func handlePutServiceRegisterData(delegate ServiceDelegate, key []byte, val []byte, leaseID int64) {
	srvtype, srvid := parseRegisterPath(key)
	var registerInfo ServiceRegisterInfo
	err := json.Unmarshal(val, &registerInfo)
	if err != nil {
		gwlog.Panic(err)
	}

	registerInfo.IsMyLease = leaseID == atomic.LoadInt64(&currentLeaseID)

	aliveServicesLock.Lock()
	defer aliveServicesLock.Unlock()
	srvtypemgr := aliveServicesByType[srvtype]
	if srvtypemgr == nil {
		srvtypemgr = newServiceTypeMgr()
		aliveServicesByType[srvtype] = srvtypemgr
	}

	srvtypemgr.registerService(srvid, registerInfo)
	gwlog.Infof("Service discoveried: %s.%s = %+v, IsMyLease %v", srvtype, srvid, registerInfo, registerInfo.IsMyLease)
	delegate.OnServiceDiscovered(srvtype, srvid, registerInfo.Addr)
}

func handleDeleteServiceRegisterData(delegate ServiceDelegate, key []byte) {
	srvtype, srvid := parseRegisterPath(key)

	aliveServicesLock.Lock()
	defer aliveServicesLock.Unlock()
	srvtypemgr := aliveServicesByType[srvtype]
	if srvtypemgr == nil {
		gwlog.Warnf("service %s.%s outdated, not not registered", srvtype, srvid)
		return
	}

	srvtypemgr.unregisterService(srvid)
	gwlog.Warnf("Service outdated: %s.%s", srvtype, srvid)
	delegate.OnServiceOutdated(srvtype, srvid)
}
