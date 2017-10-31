package main

import (
	"os"
	"os/exec"
)

func start(serverId string) {
	detectRunningServer()
	startDispatcher()
	startGates()
}

func startDispatcher() {
	var err error
	err = os.Chdir(env.GoWorldRoot)
	checkErrorOrQuit(err, "change directory failed")
	cmd := exec.Command(env.GetDispatcherExecutive())
	err = cmd.Start()
	checkErrorOrQuit(err, "start dispatcher failed")
}

func startGates() {

}
