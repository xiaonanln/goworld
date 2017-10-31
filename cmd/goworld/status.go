package main

import (
	"path/filepath"

	"strings"

	"github.com/keybase/go-ps"
	"github.com/xiaonanln/goworld/engine/config"
)

type ServerStatus struct {
	NumDispatcherRunning int
	NumGatesRunning      int
	NumGamesRunning      int
	DispatcherProcs      []ps.Process
	GateProcs            []ps.Process
	GameProcs            []ps.Process
	ServerID             ServerID
}

func (ss *ServerStatus) IsRunning() bool {
	return ss.NumDispatcherRunning > 0 || ss.NumGatesRunning > 0 || ss.NumGamesRunning > 0
}

func detectServerStatus() *ServerStatus {
	ss := &ServerStatus{}
	procs, err := ps.Processes()
	checkErrorOrQuit(err, "list processes failed")
	for _, proc := range procs {
		path, err := proc.Path()
		if err != nil {
			continue
		}

		relpath, err := filepath.Rel(env.GoWorldRoot, path)
		if err != nil || strings.HasPrefix(relpath, "..") {
			continue
		}

		dir, file := filepath.Split(relpath)

		if file == "dispatcher"+ExecutiveExt {
			ss.NumDispatcherRunning += 1
			ss.DispatcherProcs = append(ss.DispatcherProcs, proc)
		} else if file == "gate"+ExecutiveExt {
			ss.NumGatesRunning += 1
			ss.GateProcs = append(ss.GateProcs, proc)
		} else {
			ss.NumGamesRunning += 1
			ss.GameProcs = append(ss.GameProcs, proc)
			if ss.ServerID == "" {
				if strings.HasSuffix(dir, string(filepath.Separator)) {
					dir = dir[:len(dir)-1]
				}
				ss.ServerID = ServerID(strings.Join(strings.Split(dir, string(filepath.Separator)), "/"))
			}
		}
	}

	return ss
}

func status() {
	ss := detectServerStatus()
	showMsg("%d dispatcher running, %d/%d gates running, %d/%d games (%s) running", ss.NumDispatcherRunning,
		ss.NumGatesRunning, len(config.GetGateIDs()),
		ss.NumGamesRunning, len(config.GetGameIDs()),
		ss.ServerID)
}
