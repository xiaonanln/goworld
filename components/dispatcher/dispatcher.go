package main

import (
	"fmt"
	"os"
	"syscall"

	"flag"

	_ "net/http/pprof"

	"os/signal"

	"github.com/xiaonanln/goworld/components/binutil"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

var (
	configFile = ""
	sigChan    = make(chan os.Signal, 1)
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

	dispatcherConfig := config.GetDispatcher()
	binutil.SetupGWLog(dispatcherConfig.LogLevel, dispatcherConfig.LogFile, dispatcherConfig.LogStderr)
	setupSignals()
	binutil.SetupHTTPServer(dispatcherConfig.HTTPIp, dispatcherConfig.HTTPPort, nil)

	dispatcher := newDispatcherService()
	dispatcher.run()
}

func setupSignals() {
	signal.Ignore(syscall.Signal(10), syscall.Signal(12))
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			sig := <-sigChan

			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				// interrupting, quit dispatcher
				gwlog.Info("Dispatcher quited.")
				os.Exit(0)
			} else {
				gwlog.Info("unexcepted signal: %s", sig)
			}
		}
	}()
}
