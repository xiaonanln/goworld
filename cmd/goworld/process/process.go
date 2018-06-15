package process

import (
	"syscall"

	psutil_process "github.com/shirou/gopsutil/process"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type Process interface {
	Pid() int32
	Executable() string
	Path() (string, error)
	Signal(sig syscall.Signal)
}

type process struct {
	*psutil_process.Process
}

func (p process) Pid() int32 {
	return p.Process.Pid
}
func (p process) Executable() string {
	name, _ := p.Process.Name()
	return name
}

func (p process) Path() (string, error) {
	return p.Process.Exe()
}

func (p process) Signal(sig syscall.Signal) {
	p.Process.SendSignal(sig)
}

func (p process) Test() {
	name, err := p.Process.Name()
	gwlog.Infof("process %s name %s err %v", p, name, err)
}

func Processes() ([]Process, error) {
	var procs []Process

	ps, err := psutil_process.Processes()
	if err != nil {
		return nil, err
	}

	for _, _p := range ps {
		p := &process{_p}
		p.Test()
		procs = append(procs, p)
	}
	return procs, nil
}
