// +build !windows

package main

import "syscall"

const (
	IsWindows    = false
	ExecutiveExt = ""
	StopSignal   = syscall.SIGTERM
	FreezeSignal = syscall.SIGINT
)
