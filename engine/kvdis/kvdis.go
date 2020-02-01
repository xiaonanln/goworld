package kvdis

import (
	"strings"

	"github.com/xiaonanln/goworld/engine/dispatchercluster"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/post"
)

type RegisterCallback func(ok bool)

var (
	kvmap         = map[string]string{}
	postCallbacks []post.PostCallback
)

func Register(key string, val string, force bool) {
	gwlog.Infof("kvdis: register %s = %s, force=%v", key, val, force)
	dispatchercluster.SendKvdisRegister(key, val, force)
}

func TraverseByPrefix(prefix string, cb func(key string, val string)) {
	for key, val := range kvmap {
		if strings.HasPrefix(key, prefix) {
			cb(key, val)
		}
	}
}

func WatchKvdisRegister(key string, val string) {
	gwlog.Infof("kvdis: watch %s = %s", key, val)
	kvmap[key] = val

	for _, c := range postCallbacks {
		post.Post(c)
	}
}

func ClearByDispatcher(dispid uint16) {
	removeKeys := []string(nil)
	for key, _ := range kvmap {
		if dispatchercluster.SrvIDToDispatcherID(key) == dispid {
			removeKeys = append(removeKeys, key)
		}
	}
	for _, key := range removeKeys {
		delete(kvmap, key)
	}

	for _, c := range postCallbacks {
		post.Post(c)
	}
}

func AddPostCallback(cb post.PostCallback) {
	postCallbacks = append(postCallbacks, cb)
}
