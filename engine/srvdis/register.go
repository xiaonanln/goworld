package srvdis

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

var

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

	for range ch {
		gwlog.Debugf("srvdis: keep alive lease %d", lease.ID)
	}

	ctxerr := srvdisCtx.Err()
	if ctxerr == nil {
		gwlog.Panicf("srvdis: keep alive terminated, is etcd down?")
	}
}

func Register(srvtype, srvid string, info ServiceRegisterInfo) {
	//regKey := registerPath(srvtype, srvid)
	//regData := info.String()
	//srvdisKV.Txn(context.Background()).If(
	//	clientv3.Compare(clientv3.LeaseValue(regKey), "=", clientv3.NoLease),
	//).Then(
	//	clientv3.OpPut(regKey, regData, clientv3.WithLease()),
	//)
}
