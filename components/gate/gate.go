package main

import (
	"flag"

	"math/rand"
	"time"

	"os"

	_ "net/http/pprof"

	"runtime"

	"os/signal"

	"syscall"

	"github.com/xiaonanln/goworld/engine/binutil"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcherclient"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/proto"
)

var (
	gateid      uint16
	configFile  string
	logLevel    string
	gateService *GateService
	signalChan  = make(chan os.Signal, 1)
)

func init() {
	parseArgs()
}

func parseArgs() {
	var gateIdArg int
	flag.IntVar(&gateIdArg, "gid", 0, "set gateid")
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.StringVar(&logLevel, "log", "", "set log level, will override log level in config")
	flag.Parse()
	gateid = uint16(gateIdArg)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	gateConfig := config.GetGate(gateid)
	if gateConfig.GoMaxProcs > 0 {
		gwlog.Infof("SET GOMAXPROCS = %d", gateConfig.GoMaxProcs)
		runtime.GOMAXPROCS(gateConfig.GoMaxProcs)
	}
	if logLevel == "" {
		logLevel = gateConfig.LogLevel
	}
	binutil.SetupGWLog(logLevel, gateConfig.LogFile, gateConfig.LogStderr)

	gateService = newGateService()
	binutil.SetupHTTPServer(gateConfig.HTTPIp, gateConfig.HTTPPort, gateService.handleWebSocketConn)
	dispatcherclient.Initialize(&dispatcherClientDelegate{}, true)
	setupSignals()
	gateService.run() // run gate service in another goroutine
}

func setupSignals() {
	gwlog.Infof("Setup signals ...")
	signal.Ignore(syscall.Signal(10), syscall.Signal(12))
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			sig := <-signalChan
			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				// terminating gate ...
				gwlog.Infof("Terminating gate service ...")
				gateService.terminate()
				gateService.terminated.Wait()
				gwlog.Infof("Gate %d terminated gracefully.", gateid)
				os.Exit(0)
			} else {
				gwlog.Errorf("unexpected signal: %s", sig)
			}
		}
	}()
}

type dispatcherClientDelegate struct {
}

func (delegate *dispatcherClientDelegate) OnDispatcherClientConnect(dispatcherClient *dispatcherclient.DispatcherClient, isReconnect bool) {
	// called when connected / reconnected to dispatcher (not in main routine)
	dispatcherClient.SendSetGateID(gateid)
}

var lastWarnGateServiceQueueLen = 0

func (delegate *dispatcherClientDelegate) HandleDispatcherClientPacket(msgtype proto.MsgType, packet *netutil.Packet) {
	gateService.packetQueue.Push(packetQueueItem{
		msgtype: msgtype,
		packet:  packet,
	})
	qlen := gateService.packetQueue.Len()
	if qlen >= 1000 && qlen%1000 == 0 && lastWarnGateServiceQueueLen != qlen {
		gwlog.Warnf("Gate service queue length = %d", qlen)
		lastWarnGateServiceQueueLen = qlen
	}
}

func (delegate *dispatcherClientDelegate) HandleDispatcherClientDisconnect() {
	//gwlog.Errorf("Disconnected from dispatcher, try reconnecting ...")
	// if gate is disconnected from dispatcher, we just quit
	gwlog.Infof("Disconnected from dispatcher, gate has to quit.")
	signalChan <- syscall.SIGTERM // let gate quit
}

func (delegate *dispatcherClientDelegate) HandleDispatcherClientBeforeFlush() {
	gateService.handleDispatcherClientBeforeFlush()
}

// GetGateID gets the gate ID
func GetGateID() uint16 {
	return gateid
}
