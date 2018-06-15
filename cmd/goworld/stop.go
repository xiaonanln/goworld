package main

import (
	"os"
	"syscall"
	"time"

	"github.com/xiaonanln/goworld/cmd/goworld/process"
)

func stop(sid ServerID) {
	stopWithSignal(sid, StopSignal)
}

func stopWithSignal(sid ServerID, signal syscall.Signal) {
	err := os.Chdir(env.GoWorldRoot)
	checkErrorOrQuit(err, "chdir to goworld directory failed")

	ss := detectServerStatus()
	showServerStatus(ss)
	if !ss.IsRunning() {
		// server is not running
		showMsgAndQuit("no server is running currently")
	}

	if ss.ServerID != "" && ss.ServerID != sid {
		showMsgAndQuit("another server is running: %s", ss.ServerID)
	}

	stopGates(ss, signal)
	stopGames(ss, signal)
	stopDispatcher(ss, signal)
}

func stopGames(ss *ServerStatus, signal syscall.Signal) {
	if ss.NumGamesRunning == 0 {
		return
	}

	showMsg("stop %d games ...", ss.NumGamesRunning)
	for _, proc := range ss.GameProcs {
		stopProc(proc, signal)
	}
}

func stopDispatcher(ss *ServerStatus, signal syscall.Signal) {
	if ss.NumDispatcherRunning == 0 {
		return
	}

	showMsg("stop dispatcher ...")
	for _, proc := range ss.DispatcherProcs {
		stopProc(proc, signal)
	}
}

func stopGates(ss *ServerStatus, signal syscall.Signal) {
	if ss.NumGatesRunning == 0 {
		return
	}

	showMsg("stop %d gates ...", ss.NumGatesRunning)
	for _, proc := range ss.GateProcs {
		stopProc(proc, signal)
	}
}

func stopProc(proc process.Process, signal syscall.Signal) {
	showMsg("stop process %s pid=%d", proc.Executable(), proc.Pid())

	proc.Signal(signal)
	for {
		time.Sleep(time.Millisecond * 100)
		if !checkProcessRunning(proc) {
			break
		}
	}
}

func checkProcessRunning(proc process.Process) bool {
	pid := proc.Pid()
	procs, err := process.Processes()
	checkErrorOrQuit(err, "list processes failed")
	for _, _proc := range procs {
		if _proc.Pid() == pid {
			return true
		}
	}
	return false
}
