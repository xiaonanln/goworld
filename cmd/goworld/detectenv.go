package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xiaonanln/goworld/engine/config"
)

type _Env struct {
	GoWorldRoot string
}

func (env *_Env) GetDispatcherDir() string {
	return filepath.Join(env.GoWorldRoot, "components", "dispatcher")
}

func (env *_Env) GetGateDir() string {
	return filepath.Join(env.GoWorldRoot, "components", "gate")
}

func (env *_Env) GetDispatcherExecutive() string {
	return filepath.Join(env.GetDispatcherDir(), "dispatcher"+ExecutiveExt)
}

func (env *_Env) GetGateExecutive() string {
	return filepath.Join(env.GetGateDir(), "gate"+ExecutiveExt)
}

var env _Env

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

func detectGoWorldPath() {
	searchPaths := getGoSearchPaths()
	showMsg("go search paths: %s", strings.Join(searchPaths, string(os.PathListSeparator)))
	for _, sp := range searchPaths {
		goworldPath := filepath.Join(sp, "src", "github.com", "xiaonanln", "goworld")
		if isdir(goworldPath) {
			env.GoWorldRoot = goworldPath
			break
		}
	}
	if env.GoWorldRoot == "" {
		showMsgAndQuit("goworld directory is not detected")
	}

	showMsg("goworld directory found: %s", env.GoWorldRoot)
	configFile := filepath.Join(env.GoWorldRoot, "goworld.ini")
	config.SetConfigFile(configFile)
	//config.Get()
}
