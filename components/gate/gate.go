package gate

import (
	"fmt"
	"math/rand"
	_ "net/http/pprof"
	"os"
	"path"
	"runtime"
	"syscall"
	"time"

	"github.com/xiaonanln/goworld/engine/binutil"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/dispatchercluster"
	"github.com/xiaonanln/goworld/engine/dispatchercluster/dispatcherclient"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/proto"
)

// Start fires up the gate server instance
func Start() {
	rand.Seed(time.Now().UnixNano())
	parseArgs()

	if args.runInDaemonMode {
		daemoncontext := binutil.Daemonize()
		defer daemoncontext.Release()
	}

	if args.configFile != "" {
		config.SetConfigFile(args.configFile)
	}

	if args.gateid <= 0 {
		gwlog.Errorf("gateid %d is not valid, should be positive", args.gateid)
		os.Exit(1)
	}

	gateConfig := config.GetGate(args.gateid)
	verifyGateConfig(gateConfig)
	if gateConfig.GoMaxProcs > 0 {
		gwlog.Infof("SET GOMAXPROCS = %d", gateConfig.GoMaxProcs)
		runtime.GOMAXPROCS(gateConfig.GoMaxProcs)
	}
	logLevel := args.logLevel
	if logLevel == "" {
		logLevel = gateConfig.LogLevel
	}
	binutil.SetupGWLog(fmt.Sprintf("gate%d", args.gateid), logLevel, gateConfig.LogFile, gateConfig.LogStderr)

	gateService = newGateService()
	if gateConfig.EncryptConnection {
		cfgdir := config.GetConfigDir()
		rsaCert := path.Join(cfgdir, gateConfig.RSACertificate)
		rsaKey := path.Join(cfgdir, gateConfig.RSAKey)
		binutil.SetupHTTPServerTLS(gateConfig.HTTPAddr, gateService.handleWebSocketConn, rsaCert, rsaKey)
	} else {
		binutil.SetupHTTPServer(gateConfig.HTTPAddr, gateService.handleWebSocketConn)
	}

	dispatchercluster.Initialize(args.gateid, dispatcherclient.GateDispatcherClientType, false, false, &gateDispatcherClientDelegate{})
	//dispatcherclient.Initialize(&gateDispatcherClientDelegate{}, true)
	setupSignals()
	gateService.run() // run gate service in another goroutine
}

// HandleDispatcherClientPacket sends client packets to queue
func (delegate *gateDispatcherClientDelegate) HandleDispatcherClientPacket(msgtype proto.MsgType, packet *netutil.Packet) {
	gateService.dispatcherClientPacketQueue <- proto.Message{msgtype, packet}
}

// HandleDispatcherClientDisconnect sends a signal to disconnects a client
func (delegate *gateDispatcherClientDelegate) HandleDispatcherClientDisconnect() {
	//gwlog.Errorf("Disconnected from dispatcher, try reconnecting ...")
	// if gate is disconnected from dispatcher, we just quit
	gwlog.Infof("Disconnected from dispatcher, gate has to quit.")
	signalChan <- syscall.SIGTERM // let gate quit
}

// GetEntityIDsForDispatcher currently does nothing
func (deleget *gateDispatcherClientDelegate) GetEntityIDsForDispatcher(dispid uint16) []common.EntityID {
	return nil
}
