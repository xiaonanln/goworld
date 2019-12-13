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

const _Dispatcher = "dispatcher"
const _Gate = "gate"

const defaultConfigFile = "goworld.ini"

// GetSourceDir returns the path to source files
func (env *Env) GetSourceDir() string {
	return filepath.Join(env.WorkspaceRoot, "src")
}

// GetBinaryDir returns the path to binaries
func (env *Env) GetBinaryDir() string {
	return filepath.Join(env.WorkspaceRoot, "bin")
}

// GetComponentBinaryName returns executable file name of the specified component
func (env *Env) GetComponentBinaryName(component string) string {
	return component + BinaryExtension
}

// GetComponentDir returns path to the specified component
func (env *Env) GetComponentDir(component string) string {
	// We first check if there's a directory in the source tree that
	// corresponds to the specified component
	dir := filepath.Join(env.WorkspaceRoot, "components", component)
	if isexists(dir) {
		return dir
	}

	// If it's not in the source tree, the presume to be inside
	// the `goworld` directory
	return filepath.Join(env.GoWorldRoot, "components", component)
}

// GetComponentBinaryDir returns path to the specified component
func (env *Env) GetComponentBinaryDir(component string) string {
	// We'll have to iterate all possible locations of a binary.
	// Just to maintain better backward compatibility.
	dirs := []string{
		env.GetBinaryDir(),
		filepath.Join(env.WorkspaceRoot, "components"),
		filepath.Join(env.GoWorldRoot, "components"),
	}
	file := env.GetComponentBinaryName(component)
	for _, dir := range dirs {
		if isexists(filepath.Join(dir, file)) {
			return dir
		}
	}
	return ""
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
	configFile := filepath.Join(env.GetBinaryDir(), defaultConfigFile)
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

	config.SetConfigFile(configFile)
}
