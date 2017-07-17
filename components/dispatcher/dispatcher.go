package main

import (
	"fmt"
	"io"
	"os"

	"flag"

	"net/http"
	_ "net/http/pprof"

	"os/signal"

	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/gwutils"
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

func setupGWLog(dispatcherConfig *config.DispatcherConfig) {
	gwlog.Info("Set log level to %s", dispatcherConfig.LogLevel)
	gwlog.SetLevel(gwlog.StringToLevel(dispatcherConfig.LogLevel))

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

	dispatcherConfig := config.GetDispatcher()
	setupGWLog(dispatcherConfig)
	setupSignals()
	setupPprofServer(dispatcherConfig)

	dispatcher := newDispatcherService()
	dispatcher.run()
}

func setupSignals() {
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		for {
			sig := <-sigChan

			if sig == os.Interrupt {
				// interrupting, quit dispatcher
				os.Exit(0)
			} else {
				gwlog.Info("unexcepted signal: %s", sig)
			}
		}
	}()
}

func setupPprofServer(cfg *config.DispatcherConfig) {
	if cfg.PProfPort == 0 {
		// pprof not enabled
		gwlog.Info("pprof server not enabled")
		return
	}

	pprofHost := fmt.Sprintf("%s:%d", cfg.PProfIp, cfg.PProfPort)
	gwlog.Info("pprof server listening on http://%s/debug/pprof/ ... available commands: ", pprofHost)
	gwlog.Info("    go tool pprof http://%s/debug/pprof/heap", pprofHost)
	gwlog.Info("    go tool pprof http://%s/debug/pprof/profile", pprofHost)
	go func() {
		http.ListenAndServe(pprofHost, nil)
	}()
}
