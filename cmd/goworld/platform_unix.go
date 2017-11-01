// +build !windows

package main

import (
	"os"
	"syscall"
)

const (
	IsWindows    = false
	ExecutiveExt = ""
	StopSignal   = syscall.SIGTERM
	FreezeSignal = os.Interrupt
)
