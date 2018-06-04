package srvdis

import (
	"context"

	"encoding/json"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

func watchRoutine(ctx context.Context, cli *clientv3.Client, delegate ServiceDelegate) {
	kv := clientv3.NewKV(cli)
	if srvdisNamespace != "" {
		kv = namespace.NewKV(kv, srvdisNamespace)
	}

	rangeResp, err := kv.Get(ctx, "/srvdis/", clientv3.WithPrefix())
	if err != nil {
		gwlog.Fatal(err)
	}

	for _, kv := range rangeResp.Kvs {
		handlePutServiceRegisterData(delegate, kv.Key, kv.Value)
	}

	w := clientv3.NewWatcher(cli)
	if srvdisNamespace != "" {
		w = namespace.NewWatcher(w, srvdisNamespace)
	}

	ch := w.Watch(ctx, "/srvdis/", clientv3.WithPrefix(), clientv3.WithRev(rangeResp.Header.Revision+1))
	for resp := range ch {
		for _, event := range resp.Events {
			if event.Type == mvccpb.PUT {
				//gwlog.Infof("watch resp: %v, created=%v, cancelled=%v, events=%q", resp, resp.Created, resp.Canceled, resp.Events[0].Kv.Key)
				handlePutServiceRegisterData(delegate, event.Kv.Key, event.Kv.Value)
			} else if event.Type == mvccpb.DELETE {
				handleDeleteServiceRegisterData(delegate, event.Kv.Key)
			}
		}
	}
}

func handlePutServiceRegisterData(delegate ServiceDelegate, key []byte, val []byte) {
	srvtype, srvid := parseRegisterPath(key)
	var registerInfo serviceRegsiterInfo
	err := json.Unmarshal(val, &registerInfo)
	if err != nil {
		gwlog.Panic(err)
	}

	gwlog.Infof("Service discoveried: %s.%s = %s", srvtype, srvid, registerInfo)
	delegate.OnServiceDiscovered(srvtype, srvid, registerInfo.Addr)
}

func handleDeleteServiceRegisterData(delegate ServiceDelegate, key []byte) {
	srvtype, srvid := parseRegisterPath(key)
	gwlog.Warnf("Service outdated: %s.%s", srvtype, srvid)
	delegate.OnServiceOutdated(srvtype, srvid)
}
