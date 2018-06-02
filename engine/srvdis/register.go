package srvdis

import (
	"context"

	"encoding/json"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

var ()

type serviceRegsiterInfo struct {
	Addr string `json:"addr"`
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
		if err == context.Canceled || err == context.DeadlineExceeded {
			return
		}
		gwlog.Panic(errors.Wrap(err, "srvdis: grant lease failed"))
	}

	servicePath := registerPath(srvType, srvId)
	registerInfo := serviceRegsiterInfo{
		Addr: delegate.ServiceAddr(),
	}
	registerInfoBytes, err := json.Marshal(&registerInfo)
	if err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			return
		}
		gwlog.Panic(errors.Wrap(err, "srvdis: marshal register info failed"))
	}

	registerInfoStr := string(registerInfoBytes)

	gwlog.Debugf("Registering service %s = %s with lease %v TTL %v ...", servicePath, registerInfoStr, lease.ID, lease.TTL)
	_, err = kv.Put(ctx, servicePath, registerInfoStr, clientv3.WithLease(lease.ID))
	if err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			return
		}
		gwlog.Panic(errors.Wrap(err, "srvdis: etcd put failed"))
	}

	ch, err := cli.KeepAlive(ctx, lease.ID)
	if err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			return
		}
		gwlog.Panic(errors.Wrap(err, "srvdis: etcd keep alive failed"))
	}

	for resp := range ch {
		gwlog.Debugf("srvdis: service %s keep alive: %q", servicePath, resp.String())
	}

	ctxerr := ctx.Err()
	if ctxerr == nil {
		gwlog.Panicf("srvdis: %s: keep alive terminated, is etcd down?", servicePath)
	}
}
