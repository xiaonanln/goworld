package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func build(sid ServerID) {
	showMsg("building server %s ...", sid)

	buildServer(sid)
	buildComponent(_Dispatcher)
	buildComponent(_Gate)
}

func buildServer(sid ServerID) {
	serverPath := sid.Path()
	showMsg("server directory is %s ...", serverPath)
	if !isdir(serverPath) {
		showMsgAndQuit("wrong server id: %s, using '\\' instead of '/'?", sid)
	}

	showMsg("go build %s ...", sid)
	buildDirectory(serverPath)
}

// buildComponent starts the actual build process of the specified component.
// In order to support customization and use `goworld` as a pure dependency.
// We first check if there's a corresponding directory in the work space
// for the component.
// If there's no such directory in work space, we assume that it's
// using the `goworld` way, where components are inside `goworld`'s
// sub-directory.
func buildComponent(sid ServerID, component string) {
	showMsg(fmt.Sprintf("go build %s ...", component))
	buildDirectory(env.GetComponentDir(component))
}

func buildDirectory(dir string) {
	var err error

	fmt.Printf("building directory %s ...", dir)
	err = os.Chdir(dir)
	checkErrorOrQuit(err, "Failed to change directory.")

	defer func() {
		err = os.Chdir(env.WorkspaceRoot)
		checkErrorOrQuit(err, "Couldn't change to workspace directory")
	}()

	cmd := exec.Command(
		"go",
		"build",
		"-o",
		filepath.Join(binPath()),
		".",
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	checkErrorOrQuit(err, fmt.Sprintf("Failed to build %s", dir))

	// moveBinary(dir)

	return
}
