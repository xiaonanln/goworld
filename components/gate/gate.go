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

	"github.com/xiaonanln/goworld/components/binutil"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
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
		gwlog.Info("SET GOMAXPROCS = %d", gateConfig.GoMaxProcs)
		runtime.GOMAXPROCS(gateConfig.GoMaxProcs)
	}
	if logLevel == "" {
		logLevel = gateConfig.LogLevel
	}
	binutil.SetupGWLog(logLevel, gateConfig.LogFile, gateConfig.LogStderr)

	binutil.SetupPprofServer(gateConfig.PProfIp, gateConfig.PProfPort)
	dispatcher_client.Initialize(&dispatcherClientDelegate{})
	gateService = newGateService()
	setupSignals()
	gateService.run() // run gate service in another goroutine
}

func setupSignals() {
	gwlog.Info("Setup signals ...")
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			sig := <-signalChan
			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				// terminating gate ...
				gwlog.Info("Terminating gate service ...")
				gateService.terminate()
				gateService.terminated.Wait()
				gwlog.Info("Terminating game service ...")
				os.Exit(0)
			} else {
				gwlog.Error("unexpected signal: %s", sig)
			}
		}
	}()
}

type dispatcherClientDelegate struct {
}

func (delegate *dispatcherClientDelegate) OnDispatcherClientConnect(dispatcherClient *dispatcher_client.DispatcherClient, isReconnect bool) {
	// called when connected / reconnected to dispatcher (not in main routine)
	dispatcherClient.SendSetGateID(gateid)
}

var lastWarnGateServiceQueueLen = 0

func (delegate *dispatcherClientDelegate) HandleDispatcherClientPacket(msgtype proto.MsgType_t, packet *netutil.Packet) {
	gateService.packetQueue.Push(packetQueueItem{
		msgtype: msgtype,
		packet:  packet,
	})
	qlen := gateService.packetQueue.Len()
	if qlen >= 1000 && qlen%1000 == 0 && lastWarnGateServiceQueueLen != qlen {
		gwlog.Warn("Gate service queue length = %d", qlen)
		lastWarnGateServiceQueueLen = qlen
	}
}

func GetGateID() uint16 {
	return gateid
}
