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

func buildComponent(component string) {
	showMsg(fmt.Sprintf("go build %s ...", component))
	buildDirectory(env.GetComponentDir(component))
}

func buildDirectory(dir string) {
	var err error

	err = os.Chdir(dir)
	checkErrorOrQuit(err, "Failed to change directory.")

	defer func() {
		err = os.Chdir(env.WorkspaceRoot)
		checkErrorOrQuit(err, "Couldn't change to workspace directory")
	}()

	cmd := exec.Command("go", "build", ".")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	checkErrorOrQuit(err, fmt.Sprintf("Failed to build %s", dir))

	moveBinary(dir)

	return
}

func moveBinary(dir string) {
	file := filepath.Base(dir) + BinaryExtension
	fullPath := filepath.Join(dir, file)
	dest := filepath.Join(env.GetBinaryDir(), file)
	if !isfile(fullPath) {
		showMsgAndQuit("Couldn't find the binary: %s", fullPath)
	}

	err := os.Rename(fullPath, dest)
	checkErrorOrQuit(err, "Failed to move binary to workspace")

	err = os.Chmod(dest, 0544)
	checkErrorOrQuit(err, "Failed to set binary permission")
}
