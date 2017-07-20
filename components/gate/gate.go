package main

import (
	"flag"

	"math/rand"
	"time"

	"os"

	"io"

	_ "net/http/pprof"

	"net/http"

	"fmt"

	"runtime"

	"os/signal"

	"syscall"

	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/crontab"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/kvdb"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
	"github.com/xiaonanln/goworld/storage"
)

var (
	serverid    uint16
	configFile  string
	logLevel    string
	gateService *GateService
	signalChan  = make(chan os.Signal, 1)
)

func init() {
	parseArgs()
}

func parseArgs() {
	var serveridArg int
	flag.IntVar(&serveridArg, "sid", 0, "set serverid")
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.StringVar(&logLevel, "log", "", "set log level, will override log level in config")
	flag.Parse()
	serverid = uint16(serveridArg)
}

func Run(delegate IServerDelegate) {
	rand.Seed(time.Now().UnixNano())

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	serverConfig := config.GetServer(serverid)
	if serverConfig.GoMaxProcs > 0 {
		gwlog.Info("SET GOMAXPROCS = %d", serverConfig.GoMaxProcs)
		runtime.GOMAXPROCS(serverConfig.GoMaxProcs)
	}
	setupGWLog(logLevel, serverConfig)

	storage.Initialize()
	kvdb.Initialize()
	crontab.Initialize()

	setupPprofServer(serverConfig)

	entity.SetSaveInterval(serverConfig.SaveInterval)

	gameService = newGameService(serverid, delegate)

	dispatcher_client.Initialize(serverid, &dispatcherClientDelegate{})

	entity.CreateSpaceLocally(0) // create to be the nil space

	gateService = newGateService()
	go gateService.run() // run gate service in another goroutine

	setupSignals()

	gameService.run()
}

func setupSignals() {
	gwlog.Info("Setup signals ...")
	signal.Notify(signalChan, syscall.SIGINT)
	signal.Notify(signalChan, syscall.SIGTERM)

	go func() {
		for {
			sig := <-signalChan
			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				// terminating server ...
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

func setupPprofServer(serverConfig *config.ServerConfig) {
	if serverConfig.PProfPort == 0 {
		// pprof not enabled
		gwlog.Info("pprof server not enabled")
		return
	}

	pprofHost := fmt.Sprintf("%s:%d", serverConfig.PProfIp, serverConfig.PProfPort)
	gwlog.Info("pprof server listening on http://%s/debug/pprof/ ... available commands: ", pprofHost)
	gwlog.Info("    go tool pprof http://%s/debug/pprof/heap", pprofHost)
	gwlog.Info("    go tool pprof http://%s/debug/pprof/profile", pprofHost)
	go func() {
		http.ListenAndServe(pprofHost, nil)
	}()
}

func setupGWLog(logLevel string, serverConfig *config.ServerConfig) {
	if logLevel == "" {
		logLevel = serverConfig.LogLevel
	}
	gwlog.Info("Set log level to %s", logLevel)
	gwlog.SetLevel(gwlog.StringToLevel(logLevel))

	outputWriters := make([]io.Writer, 0, 2)
	if serverConfig.LogFile != "" {
		f, err := os.OpenFile(serverConfig.LogFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		outputWriters = append(outputWriters, f)
	}
	if serverConfig.LogStderr {
		outputWriters = append(outputWriters, os.Stderr)
	}

	if len(outputWriters) == 1 {
		gwlog.SetOutput(outputWriters[0])
	} else {
		gwlog.SetOutput(io.MultiWriter(outputWriters...))
	}
}

type dispatcherClientDelegate struct {
}

func (delegate *dispatcherClientDelegate) OnDispatcherClientConnect() {
	// called when connected / reconnected to dispatcher (not in main routine)
}

var lastWarnGateServiceQueueLen = 0

func (delegate *dispatcherClientDelegate) HandleDispatcherClientPacket(msgtype proto.MsgType_t, packet *netutil.Packet) {
	if msgtype >= proto.MT_GATE_SERVICE_MSG_TYPE_START && msgtype <= proto.MT_GATE_SERVICE_MSG_TYPE_STOP {
		gateService.packetQueue.Push(packetQueueItem{
			msgtype: msgtype,
			packet:  packet,
		})
		qlen := gateService.packetQueue.Len()
		if qlen >= 1000 && qlen%1000 == 0 && lastWarnGateServiceQueueLen != qlen {
			gwlog.Warn("Gate service queue length = %d", qlen)
			lastWarnGateServiceQueueLen = qlen
		}
	} else {
		gameService.packetQueue <- packetQueueItem{ // may block the dispatcher client routine
			msgtype: msgtype,
			packet:  packet,
		}
	}
}

func GetServerID() uint16 {
	return serverid
}
