// +build !windows

package main

import "syscall"

const (
	// BinaryExtension extension used on unix
	BinaryExtension = ""
	// StopSignal syscall used to stop server
	StopSignal = syscall.SIGTERM
	// FreezeSignal syscall used to freeze server
	FreezeSignal = syscall.Signal(10)
)
