package main

import (
	"flag"

	"sync"

	"math/rand"
	"time"

	_ "net/http/pprof"

	"os"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/engine/binutil"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

var (
	quiet         bool
	configFile    string
	serverHost    string
	useWebSocket  bool
	useKCP        bool
	numClients    int
	startClientId int
	noEntitySync  bool
	strictMode    bool
	duration      int
	loglevel      string
)

func parseArgs() {
	flag.BoolVar(&quiet, "quiet", false, "run client quietly with much less output")
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.IntVar(&numClients, "N", 1000, "Number of clients")
	flag.IntVar(&startClientId, "S", 1, "Start ID of clients")
	flag.StringVar(&serverHost, "server", "localhost", "replace server address")
	flag.BoolVar(&useWebSocket, "ws", false, "use WebSocket to connect server")
	flag.BoolVar(&useKCP, "kcp", false, "use KCP to connect server")
	flag.BoolVar(&noEntitySync, "nosync", false, "disable entity sync")
	flag.BoolVar(&strictMode, "strict", false, "enable strict mode")
	flag.IntVar(&duration, "duration", 0, "run for a specified duration (seconds)")
	flag.StringVar(&loglevel, "log", "info", "set log level (info by default)")
	flag.Parse()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	parseArgs()
	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	binutil.SetupGWLog("test_client", loglevel, "test_client.log", true)
	binutil.SetupHTTPServer("localhost:18888", nil)
	if useWebSocket && useKCP {
		gwlog.Errorf("Can not use both websocket and KCP")
		os.Exit(1)
	}

	if useWebSocket {
		gwlog.Infof("Using websocket clients ...")
	} else if useKCP {
		gwlog.Infof("Using KCP clients ...")
	}
	var wait sync.WaitGroup
	var waitAllConnected sync.WaitGroup
	wait.Add(numClients)
	waitAllConnected.Add(numClients)
	for i := 0; i < numClients; i++ {
		bot := newClientBot(startClientId+i, useWebSocket, useKCP, noEntitySync, &wait, &waitAllConnected)
		go bot.run()
	}
	timer.StartTicks(time.Millisecond * 100)
	if duration > 0 {
		timer.AddCallback(time.Second*time.Duration(duration), func() {
			os.Exit(0)
		})
	}
	wait.Wait()
}
