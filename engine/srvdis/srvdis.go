package srvdis

import (
	"strings"

	"github.com/xiaonanln/goworld/engine/dispatchercluster"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type RegisterCallback func(ok bool)

var (
	srvmap = map[string]string{}
)

func Register(srvid string, srvinfo string) {
	gwlog.Infof("srvdis: register %s = %s", srvid, srvinfo)
	dispatchercluster.SendSrvdisRegister(srvid, srvinfo)
}

func TraverseByPrefix(prefix string, cb func(srvid string, srvinfo string)) {
	for srvid, srvinfo := range srvmap {
		if strings.HasPrefix(srvid, prefix) {
			cb(srvid, srvinfo)
		}
	}
}

func WatchSrvdisRegister(srvid string, srvinfo string) {
	gwlog.Infof("srvdis: watch %s = %s", srvid, srvinfo)
	srvmap[srvid] = srvinfo
}
