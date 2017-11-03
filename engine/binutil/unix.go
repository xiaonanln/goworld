// +build !windows

package binutil

import (
	"os"

	"github.com/sevlyar/go-daemon"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

func Daemonize() *daemon.Context {
	context := new(daemon.Context)
	child, err := context.Reborn()

	if err != nil {
		// daemonize failed
		gwlog.Panicf("daemonize failed: %v", err)
	}

	if child != nil {
		gwlog.Infof("run in daemon mode")
		os.Exit(0)
		return nil
	} else {
		return context
	}
}
