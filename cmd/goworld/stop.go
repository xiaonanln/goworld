package main

import (
	"os"

	"syscall"

	"time"

	"github.com/keybase/go-ps"
)

func stop(serverId ServerID) {
	stopWithSignal(serverId, StopSignal)
}

func stopWithSignal(serverId ServerID, signal syscall.Signal) {
	err := os.Chdir(env.GoWorldRoot)
	checkErrorOrQuit(err, "chdir to goworld directory failed")

	ss := detectServerStatus()
	showServerStatus(ss)
	if !ss.IsRunning() {
		// server is not running
		showMsgAndQuit("no server is running currently")
	}

	if ss.ServerID != "" && ss.ServerID != serverId {
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

func stopProc(proc ps.Process, signal syscall.Signal) {
	showMsg("stop process %s pid=%d", proc.Executable(), proc.Pid())
	osproc, err := os.FindProcess(proc.Pid())
	checkErrorOrQuit(err, "stop process failed")

	osproc.Signal(signal)
	for {
		time.Sleep(time.Millisecond * 100)
		if !checkProcessRunning(proc) {
			break
		}
	}
}

func checkProcessRunning(proc ps.Process) bool {
	pid := proc.Pid()
	procs, err := ps.Processes()
	checkErrorOrQuit(err, "list processes failed")
	for _, _proc := range procs {
		if _proc.Pid() == pid {
			return true
		}
	}
	return false
}
