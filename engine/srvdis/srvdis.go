package srvdis

import (
	"time"

	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
)

var (
	srvdisNamespace string
	srvdisClient    *clientv3.Client
	srvdisKV        clientv3.KV
	srvdisCtx       context.Context
)

type ServiceDelegate interface {
	ServiceType() string
	ServiceId() string
	ServiceAddr() string
	ServiceLeaseTTL() int64
	OnServiceDiscovered(srvtype string, srvid string, srvaddr string)
	OnServiceOutdated(srvtype string, srvid string)
}

func Startup(ctx context.Context, etcdEndPoints []string, namespace_ string, delegate ServiceDelegate) {
	srvdisCtx = ctx
	srvdisNamespace = namespace_

	go func() {
		var err error
		cfg := clientv3.Config{
			Endpoints:            etcdEndPoints,
			DialTimeout:          time.Second,
			DialKeepAliveTimeout: time.Second,
			//Transport: client.DefaultTransport,
			// set timeout per request to fail fast when the target endpoint is unavailable
			//HeaderTimeoutPerRequest: time.Second,
			Context: ctx,
		}

		srvdisClient, err = clientv3.New(cfg)
		if err != nil {
			gwlog.Fatal(errors.Wrap(err, "connect etcd failed"))
		}

		defer srvdisClient.Close()

		srvdisKV = clientv3.NewKV(srvdisClient)
		if srvdisNamespace != "" {
			srvdisKV = namespace.NewKV(kv, srvdisNamespace)
		}

		go gwutils.RepeatUntilPanicless(func() {
			if ctx.Err() == nil {
				// context cancelled or exceed deadline
				registerRoutine(ctx, srvdisClient, delegate)
			}
		})

		go gwutils.RepeatUntilPanicless(func() {
			if ctx.Err() == nil {
				watchRoutine(ctx, srvdisClient, delegate)
			}
		})

		<-ctx.Done()
	}()
}
