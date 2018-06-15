// +build windows

package process

import (
	"syscall"

	"golang.org/x/sys/windows"
)

func (p process) Signal(sig syscall.Signal) {
	p.Process.SendSignal(windows.Signal(sig))
}
