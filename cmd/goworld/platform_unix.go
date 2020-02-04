// +build !windows

package main

import (
	"os"
	"syscall"
)

const (
	// BinaryExtension extension used on unix
	BinaryExtension = ""
	// StopSignal syscall used to stop server
	StopSignal = syscall.SIGTERM
)

func chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}
