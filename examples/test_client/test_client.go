package main

import (
	"flag"

	"sync"

	"math/rand"
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

var (
	quiet      bool
	configFile string
	serverAddr string
	N          int
)

func parseArgs() {
	flag.BoolVar(&quiet, "quiet", false, "run client quietly with much less output")
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.IntVar(&N, "N", 1000, "Number of clients")
	flag.StringVar(&serverAddr, "server", "localhost", "replace server address")
	flag.Parse()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	parseArgs()
	if configFile != "" {
		config.SetConfigFile(configFile)
	}
	gwlog.SetLevel(gwlog.INFO)
	var wait sync.WaitGroup
	wait.Add(N)
	for i := 0; i < N; i++ {
		bot := newClientBot(i+1, &wait)
		go bot.run()
	}
	timer.StartTicks(time.Millisecond * 100)
	wait.Wait()
}
