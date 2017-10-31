package main

import "syscall"

func kill(serverId ServerID) {
	stopWithSignal(serverId, syscall.SIGKILL)
}
