package main

import (
	"flag"

	"sync"

	"math/rand"
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/engine/binutil"
	"github.com/xiaonanln/goworld/engine/config"
)

var (
	quiet        bool
	configFile   string
	serverAddr   string
	useWebSocket bool
	numClients   int
)

func parseArgs() {
	flag.BoolVar(&quiet, "quiet", false, "run client quietly with much less output")
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.IntVar(&numClients, "N", 1000, "Number of clients")
	flag.StringVar(&serverAddr, "server", "localhost", "replace server address")
	flag.BoolVar(&useWebSocket, "ws", false, "use WebSocket to connect server")
	flag.Parse()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	parseArgs()
	if configFile != "" {
		config.SetConfigFile(configFile)
	}
	binutil.SetupGWLog("test_client", "info", "test_client.log", true)
	var wait sync.WaitGroup
	wait.Add(numClients)
	for i := 0; i < numClients; i++ {
		bot := newClientBot(i+1, useWebSocket, &wait)
		go bot.run()
	}
	timer.StartTicks(time.Millisecond * 100)
	wait.Wait()
}
