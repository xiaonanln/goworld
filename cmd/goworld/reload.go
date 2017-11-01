package main

import (
	"os"

	"github.com/xiaonanln/goworld/engine/config"
)

func reload(serverId ServerID) {
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

	if ss.NumGamesRunning == 0 {
		showMsgAndQuit("no game is running")
	} else if ss.NumGamesRunning != len(config.GetGameIDs()) {
		showMsgAndQuit("found %d games, but should have %d", ss.NumGamesRunning, len(config.GetGameIDs()))
	}

	stopGames(ss, FreezeSignal)
	startGames(serverId, true)
}
