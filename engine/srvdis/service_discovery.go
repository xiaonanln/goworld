package srvdis

import (
	"time"

	"context"

	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
)

var (
	srvdisNamespace string
)

type ServiceDelegate interface {
	ServiceType() string
	ServiceId() string
	ServiceAddr() string
	ServiceLeaseTTL() int64
	OnServiceDiscovered(srvtype string, srvid string, srvaddr string)
	OnServiceOutdated(srvtype string, srvid string)
}

func Startup(ctx context.Context, etcdEndPoints []string, namespace string, delegate ServiceDelegate) {
	go func() {
		srvdisNamespace = namespace

		cfg := clientv3.Config{
			Endpoints:            etcdEndPoints,
			DialTimeout:          time.Second,
			DialKeepAliveTimeout: time.Second,
			//Transport: client.DefaultTransport,
			// set timeout per request to fail fast when the target endpoint is unavailable
			//HeaderTimeoutPerRequest: time.Second,
			Context: ctx,
		}

		cli, err := clientv3.New(cfg)
		if err != nil {
			gwlog.Fatal(errors.Wrap(err, "connect etcd failed"))
		}

		defer cli.Close()

		var wait sync.WaitGroup
		wait.Add(2)

		go gwutils.RepeatUntilPanicless(func() {
			registerRoutine(ctx, cli, delegate)
			wait.Done()
		})

		go gwutils.RepeatUntilPanicless(func() {
			watchRoutine(ctx, cli, delegate)
			wait.Done()
		})

		wait.Wait()
	}()
}
