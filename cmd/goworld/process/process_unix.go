// +build !windows

package process

import (
	"syscall"
)

func (p process) Signal(sig syscall.Signal) {
	p.Process.SendSignal(sig)
}
