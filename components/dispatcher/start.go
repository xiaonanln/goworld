package dispatcher

import (
	"github.com/xiaonanln/goworld/engine/binutil"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// Start fires up the dispatcher server instance
func Start() {
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
