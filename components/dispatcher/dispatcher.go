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
	"github.com/xiaonanln/goworld/engine/post"
)

var (
	dispidArg         int
	dispid            uint16
	configFile        = ""
	logLevel          string
	runInDaemonMode   bool
	sigChan           = make(chan os.Signal, 1)
	dispatcherService *DispatcherService
)

func parseArgs() {
	flag.IntVar(&dispidArg, "dispid", 0, "set dispatcher ID")
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.StringVar(&logLevel, "log", "", "set log level, will override log level in config")
	flag.BoolVar(&runInDaemonMode, "d", false, "run in daemon mode")
	flag.Parse()
	dispid = uint16(dispidArg)
}

func setupGCPercent() {
	debug.SetGCPercent(consts.DISPATCHER_GC_PERCENT)
}

func main() {
	parseArgs()
	if runInDaemonMode {
		daemoncontext := binutil.Daemonize()
		defer daemoncontext.Release()
	}

	setupGCPercent()

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	validDispIds := config.GetDispatcherIDs()
	if dispid < validDispIds[0] || dispid > validDispIds[len(validDispIds)-1] {
		gwlog.Fatalf("dispatcher ID must be one of %v, but is %v, use -dispid to specify", config.GetDispatcherIDs(), dispid)
	}

	dispatcherConfig := config.GetDispatcher(dispid)

	if logLevel == "" {
		logLevel = dispatcherConfig.LogLevel
	}
	binutil.SetupGWLog("dispatcherService", logLevel, dispatcherConfig.LogFile, dispatcherConfig.LogStderr)
	binutil.SetupHTTPServer(dispatcherConfig.HTTPAddr, nil)

	dispatcherService = newDispatcherService(dispid)
	setupSignals() // call setupSignals to avoid data race on `dispatcherService`
	dispatcherService.run()
}

func setupSignals() {
	signal.Ignore(syscall.Signal(10), syscall.Signal(12), syscall.SIGPIPE, syscall.SIGHUP)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			sig := <-sigChan

			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				// interrupting, quit dispatcher
				post.Post(func() {
					dispatcherService.terminate()
				})
			} else {
				gwlog.Infof("unexcepted signal: %s", sig)
			}
		}
	}()
}
