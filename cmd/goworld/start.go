package main

import (
	"os"
	"os/exec"
)

func start(serverId string) {
	ss := detectServerStatus()
	if ss.NumDispatcherRunning > 0 || ss.NumGatesRunning > 0 {
		status()
		showMsgAndQuit("server is already running, can not start multiple servers")
	}

	err := os.Chdir(env.GoWorldRoot)
	checkErrorOrQuit(err, "chdir failed")

	startDispatcher()
	startGates()
}

func startDispatcher() {
	showMsg("start dispatcher ...")
	cmd := exec.Command(env.GetDispatcherExecutive())
	err := cmd.Start()
	checkErrorOrQuit(err, "start dispatcher failed")
}

func startGates() {
	showMsg("start gates ...")

}
