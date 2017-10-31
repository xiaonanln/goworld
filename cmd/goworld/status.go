package main

import (
	"path/filepath"

	"strings"

	"github.com/keybase/go-ps"
)

type ServerStatus struct {
	NumDispatcherRunning int
	NumGatesRunning      int
	NumGamesRunning      int
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

		//println(path)
		if !strings.Contains(path, "goworld") {
			continue
		}

		_, file := filepath.Split(path)
		println(file)
		if file == "dispatcher"+ExecutiveExt {
			ss.NumDispatcherRunning += 1
		} else if file == "gate"+ExecutiveExt {
			ss.NumGatesRunning += 1
		}
	}

	return ss
}

func status() {
	ss := detectServerStatus()
	showMsg("%d dispatcher running, %d gates running, %d games running", ss.NumDispatcherRunning, ss.NumGatesRunning, ss.NumGamesRunning)
}
