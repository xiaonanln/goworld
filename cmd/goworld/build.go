package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func build(serverId string) {
	showMsg("building server %s ...", serverId)

	buildServer(serverId)
	buildDispatcher()
	buildGate()
}

func buildServer(serverId string) {
	serverPath := getServerPath(serverId)
	showMsg("server directory is %s ...", serverPath)
	if !isdir(serverPath) {
		showMsgAndQuit("wrong server id: %s, using '\\' instead of '/'?", serverId)
	}

	showMsg("go build %s ...", serverId)
	buildDirectory(serverPath)
}

func getServerPath(serverId string) string {
	serverPath := strings.Split(serverId, "/")
	serverPath = append([]string{env.GoWorldRoot}, serverPath...)
	return filepath.Join(serverPath...)
}

func buildDispatcher() {
	showMsg("go build dispatcher ...")
	buildDirectory(filepath.Join(env.GoWorldRoot, "components", "dispatcher"))
}

func buildGate() {
	showMsg("go build gate ...")
	buildDirectory(filepath.Join(env.GoWorldRoot, "components", "gate"))
}

func buildDirectory(dir string) {
	var err error
	var curdir string
	curdir, err = os.Getwd()
	checkErrorOrQuit(err, "")

	err = os.Chdir(dir)
	checkErrorOrQuit(err, "")

	defer os.Chdir(curdir)

	cmd := exec.Command("go", "build", ".")
	cmd.Stderr = os.Stderr
	cmd.Stdout = cmd.Stdout
	cmd.Stdin = cmd.Stdin
	err = cmd.Run()
	checkErrorOrQuit(err, "")
	return
}
