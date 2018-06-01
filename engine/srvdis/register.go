package srvdis

import (
	"context"

	"fmt"

	"encoding/json"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

var ()

type serviceRegsiterInfo struct {
	Addr string `json:"addr"`
}

func registerRoutine(ctx context.Context, cli *clientv3.Client, delegate ServiceDelegate) {
	kv := clientv3.NewKV(cli)
	kv = namespace.NewKV(kv, srvdisNamespace)

	srvType := delegate.ServiceType()
	srvId := delegate.ServiceId()
	lease, err := cli.Grant(ctx, delegate.ServiceLeaseTTL())
	if err != nil {
		gwlog.Panic(err)
	}

	servicePath := registerPath(srvType, srvId)
	registerInfo := serviceRegsiterInfo{
		Addr: delegate.ServiceAddr(),
	}
	registerInfoBytes, err := json.Marshal(&registerInfo)
	if err != nil {
		gwlog.Panic(err)
	}

	registerInfoStr := string(registerInfoBytes)
	gwlog.Infof("Register info for service %s: %s", servicePath, registerInfoStr)

	gwlog.Debugf("Registering service %s with lease %v ...", servicePath, lease.ID)
	_, err = kv.Put(ctx, servicePath, "xxx", clientv3.WithLease(lease.ID))
	if err != nil {
		gwlog.Panic(err)
	}

	ch, err := cli.KeepAlive(ctx, lease.ID)
	if err != nil {
		gwlog.Panic(err)
	}

	for range ch {
		gwlog.Debugf("service %s keep alive ...", servicePath)
	}

	gwlog.Warnf("Service %s keep alive terminated.", servicePath)
}

func registerPath(srvType, srvId string) string {
	return fmt.Sprintf("/services/%s/%s", srvType, srvId)
}
