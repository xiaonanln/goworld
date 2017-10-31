package main

import (
	"os"
	"syscall"
)

func stop(serverId ServerID) {
	err := os.Chdir(env.GoWorldRoot)
	checkErrorOrQuit(err, "chdir to goworld directory failed")

	ss := detectServerStatus()
	if !ss.IsRunning() {
		// server is not running
		showMsgAndQuit("no server is running currently")
	}

	if ss.ServerID != serverId {
		showMsgAndQuit("another server is running: %s", ss.ServerID)
	}

	stopGates(ss)
}

func stopGates(ss *ServerStatus) {
	for _, proc := range ss.GateProcs {
		syscall.
	}
}
