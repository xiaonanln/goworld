package srvdis

import (
	"context"

	"fmt"

	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

var ()

func registerRoutine(ctx context.Context, client *clientv3.Client, delegate ServiceDelegate) {
	kv := clientv3.NewKV(client)
	kv = namespace.NewKV(kv, srvdisNamespace)

	srvType := delegate.ServiceType()
	srvId := delegate.ServiceId()
	lease, err := client.Grant(ctx, 5)
	if err != nil {
		gwlog.Panic(err)
	}
	leaseId := lease.ID

	for {
		gwlog.Debugf("register service %s.%s ...", srvType, srvId)

		putResp, err := kv.Put(ctx, registerPath(srvType, srvId), "xxx", clientv3.WithLease(leaseId))
		if err != nil {
			gwlog.Panic(err)
		}
		client.KeepAliveOnce(ctx, leaseId)

		gwlog.Infof("register ok: %q", putResp)
		time.Sleep(time.Second * 1)
		getresp, err := kv.Get(ctx, registerPath(srvType, srvId), clientv3.WithLimit(1))
		if err != nil {
			gwlog.Panic(err)
		}
		gwlog.Infof("registered: %s = %s", getresp.Kvs[0].Key, getresp.Kvs[0].Value)
		ttl, _ := client.TimeToLive(ctx, leaseId)
		gwlog.Infof("lease ttl: %v", ttl.TTL)

		time.Sleep(time.Second)
	}
}

func registerPath(srvType, srvId string) string {
	return fmt.Sprintf("/services/%s/%s", srvType, srvId)
}
