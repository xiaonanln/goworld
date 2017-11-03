package main

import (
	"os"
	"syscall"

	"flag"

	_ "net/http/pprof"

	"os/signal"

	"runtime/debug"

	"github.com/xiaonanln/goworld/engine/binutil"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

var (
	configFile      = ""
	logLevel        string
	runInDaemonMode bool
	sigChan         = make(chan os.Signal, 1)
)

func parseArgs() {
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.StringVar(&logLevel, "log", "", "set log level, will override log level in config")
	flag.BoolVar(&runInDaemonMode, "d", false, "run in daemon mode")
	flag.Parse()
}

func setupGCPercent() {
	debug.SetGCPercent(consts.DISPATCHER_GC_PERCENT)
}

func main() {
	if runInDaemonMode {
		daemoncontext := binutil.Daemonize()
		defer daemoncontext.Release()
	}

	setupGCPercent()
	parseArgs()

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	dispatcherConfig := config.GetDispatcher()

	if logLevel == "" {
		logLevel = dispatcherConfig.LogLevel
	}
	binutil.SetupGWLog("dispatcher", logLevel, dispatcherConfig.LogFile, dispatcherConfig.LogStderr)
	setupSignals()
	binutil.SetupHTTPServer(dispatcherConfig.HTTPIp, dispatcherConfig.HTTPPort, nil)

	dispatcher := newDispatcherService()
	dispatcher.run()
}

func setupSignals() {
	signal.Ignore(syscall.Signal(10), syscall.Signal(12), syscall.SIGPIPE)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			sig := <-sigChan

			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				// interrupting, quit dispatcher
				gwlog.Infof("Dispatcher quited.")
				os.Exit(0)
			} else {
				gwlog.Infof("unexcepted signal: %s", sig)
			}
		}
	}()
}
