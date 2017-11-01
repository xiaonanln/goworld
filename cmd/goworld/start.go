package main

import (
	"io"
	"os"
	"os/exec"

	"strconv"

	"bytes"

	"time"

	"strings"

	"path/filepath"

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
	startGames(serverId)
	startGates()
}

func startDispatcher() {
	showMsg("start dispatcher ...")
	cmd := exec.Command("nohup", env.GetDispatcherExecutive(), "~")
	err := runCmdUntilTag(cmd, consts.DISPATCHER_STARTED_TAG, time.Second*5)
	checkErrorOrQuit(err, "start dispatcher failed, see dispatcher.log for error")

}

func startGames(serverId ServerID) {
	showMsg("start games ...")
	gameIds := config.GetGameIDs()
	showMsg("game ids: %v", gameIds)
	for _, gameid := range gameIds {
		startGame(serverId, gameid)
	}
}

func startGame(serverId ServerID, gameid uint16) {
	showMsg("start game %d ...", gameid)

	gameExePath := filepath.Join(serverId.Path(), serverId.Name()+ExecutiveExt)
	cmd := exec.Command("nohup", gameExePath, "-gid", strconv.Itoa(int(gameid)))
	err := runCmdUntilTag(cmd, consts.GAME_STARTED_TAG, time.Second*5)
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

	cmd := exec.Command("nohup", env.GetGateExecutive(), "-gid", strconv.Itoa(int(gateid)))
	err := runCmdUntilTag(cmd, consts.GATE_STARTED_TAG, time.Second*5)
	checkErrorOrQuit(err, "start gate failed, see gate.log for error")
}

func runCmdUntilTag(cmd *exec.Cmd, tag string, timeout time.Duration) (err error) {
	out := bytes.NewBuffer(nil)
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, out)
	err = cmd.Start()
	if err != nil {
		return
	}
	timeoutTime := time.Now().Add(timeout)
	for time.Now().Before(timeoutTime) {
		if strings.Contains(string(out.Bytes()), "\n") {
			line, _ := out.ReadString('\n')
			//fmt.Fprintf(os.Stderr, "%s", line)
			if strings.Contains(line, tag) {
				cmd.Process.Release()
				return nil // tag received
			}
			continue
		}
		time.Sleep(time.Second)
	}

	cmd.Process.Release()
	return errors.Errorf("wait started tag timeout")
}
