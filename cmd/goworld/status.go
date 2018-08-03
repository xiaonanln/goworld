package main

import (
	"path/filepath"
	"strings"

	"fmt"

	"github.com/xiaonanln/goworld/cmd/goworld/process"
	"github.com/xiaonanln/goworld/engine/config"
)

// ServerStatus represents the status of a server
type ServerStatus struct {
	NumDispatcherRunning int
	NumGatesRunning      int
	NumGamesRunning      int

	DispatcherProcs []process.Process
	GateProcs       []process.Process
	GameProcs       []process.Process
	ServerID        ServerID
}

// IsRunning returns if a server is running
func (ss *ServerStatus) IsRunning() bool {
	return ss.NumDispatcherRunning > 0 || ss.NumGatesRunning > 0 || ss.NumGamesRunning > 0
}

func detectServerStatus() *ServerStatus {
	ss := &ServerStatus{}
	procs, err := process.Processes()
	checkErrorOrQuit(err, "list processes failed")
	for _, proc := range procs {
		path, err := proc.Path()
		if err != nil {
			continue
		}

		if !isexists(path) {
			cmdline, err := proc.CmdlineSlice()
			if err != nil {
				continue
			}
			path = cmdline[0]
			if !filepath.IsAbs(path) {
				cwd, err := proc.Cwd()
				if err != nil {
					continue
				}
				path = filepath.Join(cwd, path)
			}

		}

		relpath, err := filepath.Rel(env.GoWorldRoot, path)
		if err != nil || strings.HasPrefix(relpath, "..") {
			continue
		}

		dir, file := filepath.Split(relpath)

		if file == "dispatcher"+BinaryExtension {
			ss.NumDispatcherRunning++
			ss.DispatcherProcs = append(ss.DispatcherProcs, proc)
		} else if file == "gate"+BinaryExtension {
			ss.NumGatesRunning++
			ss.GateProcs = append(ss.GateProcs, proc)
		} else {
			if strings.HasSuffix(dir, string(filepath.Separator)) {
				dir = dir[:len(dir)-1]
			}
			serverid := ServerID(strings.Join(strings.Split(dir, string(filepath.Separator)), "/"))
			if strings.HasPrefix(string(serverid), "cmd/") || strings.HasPrefix(string(serverid), "components/") || string(serverid) == "examples/test_client" {
				// this is a cmd or a component, not a game
				continue
			}
			ss.NumGamesRunning++
			ss.GameProcs = append(ss.GameProcs, proc)
			if ss.ServerID == "" {
				ss.ServerID = serverid
			}
		}
	}

	return ss
}

func status() {
	ss := detectServerStatus()
	showServerStatus(ss)
}

func showServerStatus(ss *ServerStatus) {
	showMsg("%d dispatcher running, %d/%d gates running, %d/%d games (%s) running", ss.NumDispatcherRunning,
		ss.NumGatesRunning, config.GetDeployment().DesiredGates,
		ss.NumGamesRunning, config.GetDeployment().DesiredGames,
		ss.ServerID,
	)

	var listProcs []process.Process
	listProcs = append(listProcs, ss.DispatcherProcs...)
	listProcs = append(listProcs, ss.GameProcs...)
	listProcs = append(listProcs, ss.GateProcs...)
	for _, proc := range listProcs {
		cmdlineSlice, err := proc.CmdlineSlice()
		var cmdline string
		if err == nil {
			cmdline = strings.Join(cmdlineSlice, " ")
		} else {
			cmdline = fmt.Sprintf("get cmdline failed: %e", err)
		}

		showMsg("\t%-10d%-16s%s", proc.Pid(), proc.Executable(), cmdline)
	}
}
