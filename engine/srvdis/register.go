package srvdis

import (
	"encoding/json"

	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type registerChanItem struct {
	srvtype, srvid string
	info           ServiceRegisterInfo
}

var (
	registerChan = make(chan registerChanItem, 100)
)

type ServiceRegisterInfo struct {
	Addr string `json:"addr"`
}

func (info ServiceRegisterInfo) String() string {
	bytes, err := json.Marshal(info)
	if err != nil {
		gwlog.Panic(errors.Wrap(err, "srvdis: marshal register info failed"))
	}

	return string(bytes)
}

func registerRoutine() {
	lease, err := srvdisClient.Grant(srvdisCtx, srvdisLeaseTTL)
	if err != nil {
		gwlog.Panic(errors.Wrap(err, "srvdis: grant lease failed"))
	}

	ch, err := srvdisClient.KeepAlive(srvdisCtx, lease.ID)
	if err != nil {
		gwlog.Panic(errors.Wrap(err, "srvdis: etcd keep alive failed"))
	}

forloop:
	for {
		select {
		case resp, ok := <-ch:
			if ok {
				gwlog.Debugf("srvdis: keep alive lease %d, resp %v, ok %v", lease.ID, resp, ok)
			} else {
				break forloop
			}
			break
		case regItem := <-registerChan:
			srvtype, srvid := regItem.srvtype, regItem.srvid
			regKey := registerPath(srvtype, srvid)
			regData := regItem.info.String()
			//register the service atomically
			srvdisKV.Txn(srvdisCtx).If(
				clientv3.Compare(clientv3.LeaseValue(regKey), "=", clientv3.NoLease),
			).Then(
				clientv3.OpPut(regKey, regData, clientv3.WithLease(lease.ID)),
			).Commit()
			break
		}
	}

	ctxerr := srvdisCtx.Err()
	if ctxerr == nil {
		gwlog.Panicf("srvdis: keep alive terminated, is etcd down?")
	}
}

func Register(srvtype, srvid string, info ServiceRegisterInfo) {
	registerChan <- registerChanItem{srvtype, srvid, info}
}
