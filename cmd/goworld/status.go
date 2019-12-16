package main

import (
	"fmt"
	"strings"

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

// detectServerStatus finds out if there's any server instance running.
// Originally, this function matches only on file names, which is pretty
// risky on server environment. For now, we'll just change the checking
// to full path matches.
func detectServerStatus(sid ServerID) *ServerStatus {
	ss := &ServerStatus{}
	procs, err := process.Processes()
	checkErrorOrQuit(err, "list processes failed")
	for _, proc := range procs {
		path, err := proc.Path()
		if err != nil {
			continue
		}

		if !isexists(path) {
			// cmdline, err := proc.CmdlineSlice()
			// if err != nil {
			// 	continue
			// }
			// path = cmdline[0]
			// if !filepath.IsAbs(path) {
			// 	cwd, err := proc.Cwd()
			// 	if err != nil {
			// 		continue
			// 	}
			// 	path = filepath.Join(cwd, path)
			// }
			// this can be safely ignored now
			continue
		}

		// relpath, err := filepath.Rel(env.GoWorldRoot, path)
		// if err != nil || strings.HasPrefix(relpath, "..") {
		// 	continue
		// }
		//
		// dir, file := filepath.Split(relpath)

		switch path {
		case env.GetDispatcherBinary():
			ss.NumDispatcherRunning++
			ss.DispatcherProcs = append(ss.DispatcherProcs, proc)
		case env.GetGateBinary():
			ss.NumGatesRunning++
			ss.GateProcs = append(ss.GateProcs, proc)
		default:
			// TODO come back here to process cmd & component
			// _, base := filepath.Split(path)
			// file := filepath.Join(env.GetBinDir(), base)
			// strings.Trim(dir, string(filepath.Separator))
			// serverid := ServerID(strings.Join(strings.Split(dir, string(filepath.Separator)), "/"))
			// if strings.HasPrefix(string(serverid), "cmd/") || strings.HasPrefix(string(serverid), "components/") || string(serverid) == "examples/test_client" {
			// 	// this is a cmd or a component, not a game
			// 	continue
			// }
			if sid.BinaryPathName() == path {
				ss.NumGamesRunning++
				ss.GameProcs = append(ss.GameProcs, proc)
				if ss.ServerID == "" {
					ss.ServerID = sid // serverid
				}
			}
		}
	}

	return ss
}

func status(sid ServerID) {
	ss := detectServerStatus(sid)
	showServerStatus(ss)
}

func showServerStatus(ss *ServerStatus) {
	showMsg(
		"%d dispatcher running, %d/%d gates running, %d/%d games (%s) running",
		ss.NumDispatcherRunning,
		ss.NumGatesRunning,
		config.GetDeployment().DesiredGates,
		ss.NumGamesRunning,
		config.GetDeployment().DesiredGames,
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
