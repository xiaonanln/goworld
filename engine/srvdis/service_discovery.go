package srvdis

import (
	"time"

	"context"

	"github.com/coreos/etcd/clientv3"
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
}

func Startup(ctx context.Context, etcdEndPoints []string, namespace string, delegate ServiceDelegate) {
	srvdisNamespace = namespace

	cfg := clientv3.Config{
		Endpoints:   etcdEndPoints,
		DialTimeout: time.Second,
		//Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		//HeaderTimeoutPerRequest: time.Second,
		Context: ctx,
	}

	cli, err := clientv3.New(cfg)
	if err != nil {
		gwlog.Panic(err)
	}

	go gwutils.RepeatUntilPanicless(func() {
		registerRoutine(ctx, cli, delegate)
	})

	go gwutils.RepeatUntilPanicless(func() {
		watchRoutine(ctx, cli, delegate)
	})
}
