package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xiaonanln/goworld/engine/config"
)

// Env represents environment variables
type Env struct {
	GoWorldRoot   string
	WorkspaceRoot string
}

const defaultConfigFile = "goworld.ini"

// GetDispatcherDir returns the path to the dispatcher
func (env *Env) GetDispatcherDir() string {
	return filepath.Join(env.GoWorldRoot, "components", "dispatcher")
}

// GetGateDir returns the path to the gate
func (env *Env) GetGateDir() string {
	return filepath.Join(env.GoWorldRoot, "components", "gate")
}

// GetDispatcherBinary returns the path to the dispatcher binary
func (env *Env) GetDispatcherBinary() string {
	return filepath.Join(env.GetDispatcherDir(), "dispatcher"+BinaryExtension)
}

// GetGateBinary returns the path to the gate binary
func (env *Env) GetGateBinary() string {
	return filepath.Join(env.GetGateDir(), "gate"+BinaryExtension)
}

var env Env

func getGoSearchPaths() []string {
	var paths []string
	goroot := os.Getenv("GOROOT")
	if goroot != "" {
		paths = append(paths, goroot)
	}

	gopath := os.Getenv("GOPATH")
	for _, p := range strings.Split(gopath, string(os.PathListSeparator)) {
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}

// ModuleInfo holds information about Go modules
type ModuleInfo struct {
	Path      string `json:"Path"`
	Main      bool   `json:"Main"`
	Dir       string `json:"Dir"`
	GoMod     string `json:"GoMod"`
	GoVersion string `json:"GoVersion"`
}

func goListModule() (*ModuleInfo, error) {
	cmd := exec.Command("go", "list", "-m", "-json")

	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	d := json.NewDecoder(r)
	var mi ModuleInfo
	err = d.Decode(&mi)
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	return &mi, err
}

func _detectGoWorldPath() string {
	mi, err := goListModule()
	if err == nil {
		showMsg("go list -m -json: %+v", *mi)
		return mi.Dir
	}

	searchPaths := getGoSearchPaths()
	showMsg("go search paths: %s", strings.Join(searchPaths, string(os.PathListSeparator)))
	for _, sp := range searchPaths {
		goworldPath := filepath.Join(sp, "src", "github.com", "xiaonanln", "goworld")
		if isdir(goworldPath) {
			return goworldPath
		}
	}
	return ""
}

func detectGoWorldPath() {
	wsRoot, err := os.Getwd()
	checkErrorOrQuit(err, "Failed to get current work directory.")

	env.WorkspaceRoot = wsRoot
	env.GoWorldRoot = _detectGoWorldPath()
	if env.GoWorldRoot == "" {
		showMsgAndQuit("goworld directory is not detected")
	}

	showMsg("goworld directory found: %s", env.GoWorldRoot)

	checkConfigFile()
}

func checkConfigFile() {
	configFile := filepath.Join(binPath(), defaultConfigFile)
	if !isexists(configFile) {
		def := filepath.Join(env.GoWorldRoot, defaultConfigFile)
		if isexists(def) {
			err := os.Rename(def, configFile)
			checkErrorOrQuit(err, "Failed to create default config file")
		} else {
			def = filepath.Join(env.GoWorldRoot, defaultConfigFile+".sample")
			err := copyFile(def, configFile)
			checkErrorOrQuit(err, "Failed to create default config file")
		}

		err := os.Chmod(configFile, 0644)
		checkErrorOrQuit(err, "Failed to set config file permission")
	}

	config.SetConfigFile(configFile)
}
