// +build !windows

package main

import "syscall"

const (
	// BinaryExtension extension used on unix
	BinaryExtension = ""
	// StopSignal syscall used to stop server
	StopSignal = syscall.SIGTERM
)
