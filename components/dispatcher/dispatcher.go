package main

import (
	"fmt"

	"flag"

	"os"

	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
)

var (
	configFile = ""
)

func debuglog(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	gwlog.Debug("dispatcher: %s", s)
}

func parseArgs() {
	flag.Parse()
}

func main() {
	cfg := config.GetDispatcher()
	fmt.Fprintf(os.Stderr, "Read dispatcher config: \n%s\n", config.DumpPretty(cfg))
	//host := fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port)
	dispatcher := newDispatcherService(cfg)
	dispatcher.run()
}
