package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
)

func start(sid ServerID) {
	err := os.Chdir(env.GoWorldRoot)
	checkErrorOrQuit(err, "chdir to goworld directory failed")

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
	cfg := config.GetDispatcher(dispid)
	args := []string{}
	if arguments.runInDaemonMode {
		args = append(args, "-d")
	}
	cmd := exec.Command(env.GetDispatcherBinary(), args...)
	err := runCmdUntilTag(cmd, cfg.LogFile, consts.DISPATCHER_STARTED_TAG, time.Second*10)
	checkErrorOrQuit(err, "start dispatcher failed, see dispatcher.log for error")
}

func startGames(sid ServerID, isRestore bool) {
	showMsg("start games ...")
	gameIds := config.GetGameIDs()
	showMsg("game ids: %v", gameIds)
	for _, gameid := range gameIds {
		startGame(sid, gameid, isRestore)
	}
}

func startGame(sid ServerID, gameid uint16, isRestore bool) {
	showMsg("start game %d ...", gameid)

	gameExePath := filepath.Join(sid.Path(), sid.Name()+BinaryExtension)
	args := []string{"-gid", strconv.Itoa(int(gameid))}
	if isRestore {
		args = append(args, "-restore")
	}
	if arguments.runInDaemonMode {
		args = append(args, "-d")
	}
	cmd := exec.Command(gameExePath, args...)
	err := runCmdUntilTag(cmd, config.GetGame(gameid).LogFile, consts.GAME_STARTED_TAG, time.Second*10)
	checkErrorOrQuit(err, "start game failed, see game.log for error")
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

	args := []string{"-gid", strconv.Itoa(int(gateid))}
	if arguments.runInDaemonMode {
		args = append(args, "-d")
	}
	cmd := exec.Command(env.GetGateBinary(), args...)
	err := runCmdUntilTag(cmd, config.GetGate(gateid).LogFile, consts.GATE_STARTED_TAG, time.Second*10)
	checkErrorOrQuit(err, "start gate failed, see gate.log for error")
}

func runCmdUntilTag(cmd *exec.Cmd, logFile string, tag string, timeout time.Duration) (err error) {
	err = cmd.Start()
	if err != nil {
		return
	}

	timeoutTime := time.Now().Add(timeout)
	for time.Now().Before(timeoutTime) {
		time.Sleep(time.Millisecond * 200)
		if isTagInFile(logFile, tag) {
			cmd.Process.Release()
			return
		}
	}

	err = errors.Errorf("wait started tag timeout")
	return
}

func isTagInFile(filename string, tag string) bool {
	data, err := ioutil.ReadFile(filename)
	checkErrorOrQuit(err, "read file error")
	return strings.Contains(string(data), tag)
}
