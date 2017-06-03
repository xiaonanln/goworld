package goworld_startup

import "github.com/xiaonanln/goworld/netutil"

func Startup() {
	goworld_netutil.ServeTCPForever()
}
