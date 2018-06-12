// +build windows

package main

import (
	"syscall"

	_ "github.com/go-ole/go-ole" // so that dep can resolve versions correctly
)

const (
	// BinaryExtension extension used on windows
	BinaryExtension = ".exe"
	// StopSignal syscall used to stop server
	StopSignal = syscall.SIGKILL
)
