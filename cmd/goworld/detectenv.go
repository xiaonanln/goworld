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
	ConfigPath    string
	GoWorldRoot   string
	WorkspaceRoot string
}

const _Dispatcher = "dispatcher"
const _Gate = "gate"

const defaultConfigFile = "goworld.ini"

// GetSrcDir returns absolute path to the source directory
func (env *Env) GetSrcDir() string {
	return filepath.Join(env.WorkspaceRoot, "src")
}

// GetBinDir returns absolute path to the binary directory
func (env *Env) GetBinDir() string {
	return filepath.Join(env.WorkspaceRoot, "bin")
}

// GetServerComponentDir returns absolute path to the specified component
func (env *Env) GetServerComponentDir(sid ServerID, component string) (dir string) {
	dir = filepath.Join(sid.Path(), "components", component)
	if isexists(dir) {
		return
	}

	dir = filepath.Join(env.GoWorldRoot, "components", component)
	return
}

// GetServerComponentBinDir returns absolute path to the specified component binary
func (env *Env) GetServerComponentBinDir(component string) (dir string) {
	dir = env.GetBinDir()
	if isexists(filepath.Join(dir, component+BinaryExtension)) {
		return
	}

	dir = filepath.Join(
		env.GoWorldRoot,
		"components",
		component,
		component+BinaryExtension,
	)
	return
}

// GetDispatcherDir returns the path to the dispatcher
func (env *Env) GetDispatcherDir(sid ServerID) string {
	return env.GetServerComponentDir(sid, _Dispatcher)
}

// GetDispatcherBinary returns the path to the dispatcher binary
func (env *Env) GetDispatcherBinary() string {
	return filepath.Join(
		env.GetServerComponentBinDir(_Dispatcher),
		_Dispatcher+BinaryExtension,
	)
}

// GetGateBinary returns the path to the gate binary
func (env *Env) GetGateBinary() string {
	return filepath.Join(
		env.GetServerComponentBinDir(_Gate),
		_Gate+BinaryExtension,
	)
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
	// There's a problem while running `goworld` using debugger in `GoLand`.
	// It's possible that `go list -m` returns the following for `Path`.
	// So we have to exclude the case.
	if err == nil && mi.Path != "command-line-arguments" {
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
	configFile := filepath.Join(env.GetBinDir(), defaultConfigFile)
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

		err := chmod(configFile, 0644)
		checkErrorOrQuit(err, "Failed to set config file permission")
	}

	env.ConfigPath = configFile
	config.SetConfigFile(configFile)
}
