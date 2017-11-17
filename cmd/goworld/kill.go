package main

import "syscall"

func kill(sid ServerID) {
	stopWithSignal(sid, syscall.SIGKILL)
}
