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

	"fmt"

	"path"

	"github.com/xiaonanln/goworld/engine/binutil"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/dispatchercluster"
	"github.com/xiaonanln/goworld/engine/dispatchercluster/dispatcherclient"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/post"
	"github.com/xiaonanln/goworld/engine/proto"
)

var (
	gateid          uint16
	configFile      string
	logLevel        string
	runInDaemonMode bool
	gateService     *GateService
	signalChan      = make(chan os.Signal, 1)
)

func parseArgs() {
	var gateIdArg int
	flag.IntVar(&gateIdArg, "gid", 0, "set gateid")
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.StringVar(&logLevel, "log", "", "set log level, will override log level in config")
	flag.BoolVar(&runInDaemonMode, "d", false, "run in daemon mode")
	flag.Parse()
	gateid = uint16(gateIdArg)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	parseArgs()

	if runInDaemonMode {
		daemoncontext := binutil.Daemonize()
		defer daemoncontext.Release()
	}

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	if gateid <= 0 {
		gwlog.Errorf("gateid %d is not valid, should be positive", gateid)
		os.Exit(1)
	}

	gateConfig := config.GetGate(gateid)
	if gateConfig == nil {
		gwlog.Errorf("gate %d's config is not found", gateid)
		os.Exit(1)
	}
	if gateConfig.GoMaxProcs > 0 {
		gwlog.Infof("SET GOMAXPROCS = %d", gateConfig.GoMaxProcs)
		runtime.GOMAXPROCS(gateConfig.GoMaxProcs)
	}
	if logLevel == "" {
		logLevel = gateConfig.LogLevel
	}
	binutil.SetupGWLog(fmt.Sprintf("gate%d", gateid), logLevel, gateConfig.LogFile, gateConfig.LogStderr)

	gateService = newGateService()
	if gateConfig.EncryptConnection {
		cfgdir := config.GetConfigDir()
		rsaCert := path.Join(cfgdir, gateConfig.RSACertificate)
		rsaKey := path.Join(cfgdir, gateConfig.RSAKey)
		binutil.SetupHTTPServerTLS(gateConfig.HTTPIp, gateConfig.HTTPPort, gateService.handleWebSocketConn, rsaCert, rsaKey)
	} else {
		binutil.SetupHTTPServer(gateConfig.HTTPIp, gateConfig.HTTPPort, gateService.handleWebSocketConn)
	}

	dispatchercluster.Initialize(gateid, dispatcherclient.GateDispatcherClientType, false, false, &gateDispatcherClientDelegate{})
	//dispatcherclient.Initialize(&gateDispatcherClientDelegate{}, true)
	setupSignals()
	gateService.run() // run gate service in another goroutine
}

func setupSignals() {
	gwlog.Infof("Setup signals ...")
	signal.Ignore(syscall.Signal(10), syscall.Signal(12), syscall.SIGPIPE, syscall.SIGHUP)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			sig := <-signalChan
			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				// terminating gate ...
				gwlog.Infof("Terminating gate service ...")
				post.Post(func() {
					gateService.terminate()
				})

				gateService.terminated.Wait()
				gwlog.Infof("Gate %d terminated gracefully.", gateid)
				os.Exit(0)
			} else {
				gwlog.Errorf("unexpected signal: %s", sig)
			}
		}
	}()
}

type gateDispatcherClientDelegate struct {
}

func (delegate *gateDispatcherClientDelegate) HandleDispatcherClientPacket(msgtype proto.MsgType, packet *netutil.Packet) {
	gateService.dispatcherClientPacketQueue <- proto.Message{msgtype, packet}
}

func (delegate *gateDispatcherClientDelegate) HandleDispatcherClientDisconnect() {
	//gwlog.Errorf("Disconnected from dispatcher, try reconnecting ...")
	// if gate is disconnected from dispatcher, we just quit
	gwlog.Infof("Disconnected from dispatcher, gate has to quit.")
	signalChan <- syscall.SIGTERM // let gate quit
}

func (deleget *gateDispatcherClientDelegate) GetEntityIDsForDispatcher(dispid uint16) []common.EntityID {
	return nil
}

// GetGateID gets the gate ID
func GetGateID() uint16 {
	return gateid
}
