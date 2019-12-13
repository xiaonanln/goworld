package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
)

func start(sid ServerID) {
	ss := detectServerStatus()
	if ss.NumDispatcherRunning > 0 || ss.NumGatesRunning > 0 {
		status()
		showMsgAndQuit("server is already running, can not start multiple servers")
	}

	startDispatchers()
	startGames(sid, false)
	startGates()
}

func startDispatchers() {
	showMsg("start dispatchers ...")
	dispatcherIds := config.GetDispatcherIDs()
	showMsg("dispatcher ids: %v", dispatcherIds)
	for _, dispid := range dispatcherIds {
		startDispatcher(dispid)
	}
}

func startDispatcher(dispid uint16) {
	dir := env.GetComponentBinaryDir(_Dispatcher)
	if "" == dir {
		showMsgAndQuit("Failed to locate dispatcher executable")
	}
	err := os.Chdir(dir)
	checkErrorOrQuit(err, "Failed to change to dispatcher's directory")

	cfg := config.GetDispatcher(dispid)
	args := []string{"-dispid", strconv.Itoa(int(dispid))}
	if arguments.runInDaemonMode {
		args = append(args, "-d")
	}
	cmd := exec.Command(env.GetComponentBinaryName(_Dispatcher), args...)
	err = runCmdUntilTag(cmd, cfg.LogFile, consts.DISPATCHER_STARTED_TAG, time.Second*10)
	checkErrorOrQuit(err, "start dispatcher failed, see dispatcher.log for error")
}

func startGames(sid ServerID, isRestore bool) {
	showMsg("start games ...")
	desiredGames := config.GetDeployment().DesiredGames
	showMsg("desired games = %d", desiredGames)
	for gameid := uint16(1); int(gameid) <= desiredGames; gameid++ {
		startGame(sid, gameid, isRestore)
	}
}

func startGame(sid ServerID, gameid uint16, isRestore bool) {
	showMsg("start game %d ...", gameid)

	gameExePath := sid.BinaryPathName()
	args := []string{"-gid", strconv.Itoa(int(gameid))}
	if isRestore {
		args = append(args, "-restore")
	}
	if arguments.runInDaemonMode {
		args = append(args, "-d")
	}
	cmd := exec.Command(gameExePath, args...)
	err := runCmdUntilTag(cmd, config.GetGame(gameid).LogFile, consts.GAME_STARTED_TAG, time.Second*600)
	checkErrorOrQuit(err, "start game failed, see game.log for error")
}

func startGates() {
	showMsg("start gates ...")
	desiredGates := config.GetDeployment().DesiredGates
	showMsg("desired gates = %d", desiredGates)
	for gateid := uint16(1); int(gateid) <= desiredGates; gateid++ {
		startGate(gateid)
	}
}

func startGate(gateid uint16) {
	showMsg("start gate %d ...", gateid)

	dir := env.GetComponentBinaryDir(_Gate)
	if "" == dir {
		showMsgAndQuit("Failed to locate gate executable")
	}
	err := os.Chdir(dir)
	checkErrorOrQuit(err, "Failed to change to gate's directory")

	args := []string{"-gid", strconv.Itoa(int(gateid))}
	if arguments.runInDaemonMode {
		args = append(args, "-d")
	}
	cmd := exec.Command(env.GetComponentBinaryName(_Gate), args...)
	err = runCmdUntilTag(cmd, config.GetGate(gateid).LogFile, consts.GATE_STARTED_TAG, time.Second*10)
	checkErrorOrQuit(err, "start gate failed, see gate.log for error")
}

func runCmdUntilTag(cmd *exec.Cmd, logFile string, tag string, timeout time.Duration) (err error) {
	clearLogFile(logFile)
	err = cmd.Start()
	if err != nil {
		return
	}

	timeoutTime := time.Now().Add(timeout)
	for time.Now().Before(timeoutTime) {
		time.Sleep(time.Millisecond * 200)
		if isTagInFile(logFile, tag) {
			err = cmd.Process.Release()
			return
		}
	}

	err = errors.Errorf("wait started tag timeout")
	return
}

func clearLogFile(logFile string) {
	err := ioutil.WriteFile(logFile, []byte{}, 0644)
	if err != nil {
		fmt.Printf("Failed to write log: %v", err)
	}
}

func isTagInFile(filename string, tag string) bool {
	data, err := ioutil.ReadFile(filename)
	checkErrorOrQuit(err, "read file error")
	return strings.Contains(string(data), tag)
}
