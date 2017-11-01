// +build windows

package main

import (
	"syscall"

	_ "github.com/go-ole/go-ole" // so that dep can resolve versions correctly
)

const (
	IsWindows    = true
	ExecutiveExt = ".exe"
	StopSignal   = syscall.SIGKILL
	FreezeSignal = syscall.SIGINT
)
