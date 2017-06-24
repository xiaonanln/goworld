package main

import (
	"fmt"

	"flag"

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
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.Parse()
}

func main() {
	parseArgs()

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	//	f, err := os.OpenFile("dispatcher.log", os.O_APPEND|os.O_CREATE, 0644)
	//	if err != nil {
	//		panic(err)
	//	}
	//	gwlog.SetOutput(f)

	//cfg := config.GetDispatcher()
	//fmt.Fprintf(os.Stderr, "Read dispatcher config: \n%s\n", config.DumpPretty(cfg))
	//host := fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port)
	dispatcher := newDispatcherService()
	dispatcher.run()
}
