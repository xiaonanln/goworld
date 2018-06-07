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
	srvdisLeaseTTL  int64
	serviceDelegate ServiceDelegate
	currentLeaseID  int64
)

type ServiceDelegate interface {
	OnServiceDiscovered(srvtype string, srvid string, srvaddr string)
	OnServiceOutdated(srvtype string, srvid string)
}

func Startup(ctx context.Context, etcdEndPoints []string, namespace_ string, leaseTTL int64, delegate ServiceDelegate) {
	srvdisCtx = ctx
	srvdisNamespace = namespace_
	srvdisLeaseTTL = leaseTTL
	serviceDelegate = delegate

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
			srvdisKV = namespace.NewKV(srvdisKV, srvdisNamespace)
		}

		go gwutils.RepeatUntilPanicless(srvdisCtx, registerRoutine)
		go gwutils.RepeatUntilPanicless(srvdisCtx, watchRoutine)

		<-srvdisCtx.Done()
	}()
}
