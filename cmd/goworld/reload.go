package main

import (
	"os"

	"github.com/xiaonanln/goworld/engine/binutil"
	"github.com/xiaonanln/goworld/engine/config"
)

func reload(sid ServerID) {
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

	if ss.NumGamesRunning == 0 {
		showMsgAndQuit("no game is running")
	} else if ss.NumGamesRunning != config.GetDeployment().DesiredGames {
		showMsgAndQuit("found %d games, but should have %d", ss.NumGamesRunning, config.GetDeployment().DesiredGames)
	}

	stopGames(ss, binutil.FreezeSignal)
	startGames(sid, true)
}
