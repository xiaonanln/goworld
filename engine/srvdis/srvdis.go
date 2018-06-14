package srvdis

import (
	"strings"

	"github.com/xiaonanln/goworld/engine/dispatchercluster"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/post"
)

type RegisterCallback func(ok bool)

var (
	srvmap        = map[string]string{}
	postCallbacks []post.PostCallback
)

func Register(srvid string, srvinfo string, force bool) {
	gwlog.Infof("srvdis: register %s = %s, force=%v", srvid, srvinfo, force)
	dispatchercluster.SendSrvdisRegister(srvid, srvinfo, force)
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

	for _, c := range postCallbacks {
		post.Post(c)
	}
}

func ClearByDispatcher(dispid uint16) {
	removeSrvIDs := []string(nil)
	for srvid, _ := range srvmap {
		if dispatchercluster.SrvIDToDispatcherID(srvid) == dispid {
			removeSrvIDs = append(removeSrvIDs, srvid)
		}
	}
	for _, srvid := range removeSrvIDs {
		delete(srvmap, srvid)
	}

	for _, c := range postCallbacks {
		post.Post(c)
	}
}

func AddPostCallback(cb post.PostCallback) {
	postCallbacks = append(postCallbacks, cb)
}
