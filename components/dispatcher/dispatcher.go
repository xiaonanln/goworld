package main

import (
	"fmt"
	"io"
	"os"

	"flag"

	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/gwutils"
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

func setupLogOutput(dispatcherConfig *config.DispatcherConfig) {
	outputWriters := make([]io.Writer, 0, 2)
	if dispatcherConfig.LogFile != "" {
		f, err := os.OpenFile(dispatcherConfig.LogFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		outputWriters = append(outputWriters, f)
	}
	if dispatcherConfig.LogStderr {
		outputWriters = append(outputWriters, os.Stderr)
	}

	if len(outputWriters) == 1 {
		gwlog.SetOutput(outputWriters[0])
	} else {
		gwlog.SetOutput(gwutils.NewMultiWriter(outputWriters...))
	}
}

func main() {
	parseArgs()

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	setupLogOutput(config.GetDispatcher())
	//cfg := config.GetDispatcher()
	//fmt.Fprintf(os.Stderr, "Read dispatcher config: \n%s\n", config.DumpPretty(cfg))
	//host := fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port)
	dispatcher := newDispatcherService()
	dispatcher.run()
}
