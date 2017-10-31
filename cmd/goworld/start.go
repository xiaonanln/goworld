package main

import (
	"os"
	"os/exec"

	"strconv"

	"bytes"

	"time"

	"strings"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
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
	//startGames()
	startGates()
}

func startDispatcher() {
	showMsg("start dispatcher ...")
	cmd := exec.Command(env.GetDispatcherExecutive())
	err := runCmdUntilTag(cmd, consts.DISPATCHER_STARTED_TAG, time.Second*5)
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
	err := runCmdUntilTag(cmd, consts.GATE_STARTED_TAG, time.Second*5)
	checkErrorOrQuit(err, "start gate failed")
}

func runCmdUntilTag(cmd *exec.Cmd, tag string, timeout time.Duration) (err error) {
	out := bytes.NewBuffer(nil)
	cmd.Stderr = out
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
				return nil // tag received
			}
			continue
		}
		time.Sleep(time.Second)
	}

	return errors.Errorf("wait started tag timeout")
}
