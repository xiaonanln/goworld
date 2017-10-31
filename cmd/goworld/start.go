package main

import (
	"os"
	"os/exec"

	"strconv"

	"github.com/xiaonanln/goworld/engine/config"
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
	startGames()
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
	gateIds := config.GetGateIDs()
	showMsg("gate ids: %v", gateIds)
	for _, gateid := range gateIds {
		startGate(gateid)
	}
}

func startGate(gateid uint16) {
	showMsg("start gate %d ...", gateid)

	cmd := exec.Command(env.GetGateExecutive(), "-gid", strconv.Itoa(int(gateid)))
	err := cmd.Start()
	checkErrorOrQuit(err, "start gate failed")
}
