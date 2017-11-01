package main

import (
	"os"
	"os/exec"

	"strconv"

	"time"

	"path/filepath"

	"io/ioutil"

	"strings"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
)

func start(serverId ServerID) {
	err := os.Chdir(env.GoWorldRoot)
	checkErrorOrQuit(err, "chdir to goworld directory failed")

	ss := detectServerStatus()
	if ss.NumDispatcherRunning > 0 || ss.NumGatesRunning > 0 {
		status()
		showMsgAndQuit("server is already running, can not start multiple servers")
	}

	startDispatcher()
	startGames(serverId, false)
	startGates()
}

func startDispatcher() {
	showMsg("start dispatcher ...")
	cmd := exec.Command(env.GetDispatcherExecutive())
	err := runCmdUntilTag(cmd, config.GetDispatcher().LogFile, consts.DISPATCHER_STARTED_TAG, time.Second*10)
	checkErrorOrQuit(err, "start dispatcher failed, see dispatcher.log for error")

}

func startGames(serverId ServerID, isRestore bool) {
	showMsg("start games ...")
	gameIds := config.GetGameIDs()
	showMsg("game ids: %v", gameIds)
	for _, gameid := range gameIds {
		startGame(serverId, gameid, isRestore)
	}
}

func startGame(serverId ServerID, gameid uint16, isRestore bool) {
	showMsg("start game %d ...", gameid)

	gameExePath := filepath.Join(serverId.Path(), serverId.Name()+ExecutiveExt)
	args := []string{"-gid", strconv.Itoa(int(gameid))}
	if isRestore {
		args = append(args, "-restore")
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

	cmd := exec.Command(env.GetGateExecutive(), "-gid", strconv.Itoa(int(gateid)))
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
