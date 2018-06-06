package srvdis

import (
	"context"

	"encoding/json"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"encoding/gob"
)

var ()

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

func registerRoutine(ctx context.Context, cli *clientv3.Client, delegate ServiceDelegate) {
	kv := clientv3.NewKV(cli)
	if srvdisNamespace != "" {
		kv = namespace.NewKV(kv, srvdisNamespace)
	}

	srvType := delegate.ServiceType()
	srvId := delegate.ServiceId()
	lease, err := cli.Grant(ctx, delegate.ServiceLeaseTTL())
	if err != nil {
		gwlog.Panic(errors.Wrap(err, "srvdis: grant lease failed"))
	}

	servicePath := registerPath(srvType, srvId)
	registerInfo := ServiceRegisterInfo{
		Addr: delegate.ServiceAddr(),
	}
	registerInfoStr := registerInfo.String()

	gwlog.Debugf("Registering service %s = %s with lease %v TTL %v ...", servicePath, registerInfoStr, lease.ID, lease.TTL)
	_, err = kv.Put(ctx, servicePath, registerInfoStr, clientv3.WithLease(lease.ID))
	if err != nil {
		gwlog.Panic(errors.Wrap(err, "srvdis: etcd put failed"))
	}

	ch, err := cli.KeepAlive(ctx, lease.ID)
	if err != nil {
		gwlog.Panic(errors.Wrap(err, "srvdis: etcd keep alive failed"))
	}

	for range ch {
		//gwlog.Debugf("srvdis: service %s keep alive: %q", servicePath, resp.String())
	}

	ctxerr := ctx.Err()
	if ctxerr == nil {
		gwlog.Panicf("srvdis: %s: keep alive terminated, is etcd down?", servicePath)
	}
}

func Register(srvtype, srvid string, info ServiceRegisterInfo) {
	regKey := registerPath(srvtype, srvid)
	regData := info.String()
	srvdisKV.Txn(context.Background()).If(
		clientv3.Compare(clientv3.LeaseValue(regKey), "=", clientv3.NoLease),
	).Then(
		clientv3.OpPut(regKey, regData, )
	)
}
