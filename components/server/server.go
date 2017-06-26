package server

import (
	"flag"

	"math/rand"
	"time"

	"os"

	"io"

	_ "net/http/pprof"

	"net/http"

	"fmt"

	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/config"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/gwutils"
	"github.com/xiaonanln/goworld/kvdb"
	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/proto"
	"github.com/xiaonanln/goworld/storage"
)

var (
	serverid    uint16
	configFile  string
	gameService *GameService
	gateService *GateService
)

func init() {
	parseArgs()
}

func parseArgs() {
	var serveridArg int
	flag.IntVar(&serveridArg, "sid", 0, "set serverid")
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.Parse()
	serverid = uint16(serveridArg)
}

func Run(delegate IServerDelegate) {
	rand.Seed(time.Now().UnixNano())

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	serverConfig := config.GetServer(serverid)
	setupLogOutput(serverConfig)

	storage.Initialize()
	kvdb.Initialize()

	setupPprofServer(serverConfig)

	entity.SetSaveInterval(serverConfig.SaveInterval)

	gameService = newGameService(serverid, delegate)

	dispatcher_client.Initialize(serverid, &dispatcherClientDelegate{})

	entity.CreateSpaceLocally(0) // create to be the nil space

	gateService = newGateService()
	go gateService.run() // run gate service in another goroutine

	gameService.run()
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

func setupLogOutput(serverConfig *config.ServerConfig) {
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
		gwlog.SetOutput(gwutils.NewMultiWriter(outputWriters...))
	}
}

type dispatcherClientDelegate struct {
}

func (delegate *dispatcherClientDelegate) OnDispatcherClientConnect() {
	// called when connected / reconnected to dispatcher (not in main routine)

}
func (delegate *dispatcherClientDelegate) HandleDispatcherClientPacket(msgtype proto.MsgType_t, packet *netutil.Packet) {
	if msgtype < proto.MT_GATE_SERVICE_MSG_TYPE_START {
		gameService.packetQueue <- packetQueueItem{ // may block the dispatcher client routine
			msgtype: msgtype,
			pkt:     packet,
		}
	} else {
		gateService.HandleDispatcherClientPacket(msgtype, packet)
	}
}

func GetServerID() uint16 {
	return serverid
}
