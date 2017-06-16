package main

import (
	"flag"

	"sync"

	"math/rand"
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/config"
)

const (
	NR_CLIENTS = 1
)

var (
	configFile string
)

func parseArgs() {
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.Parse()
}

func main() {
	rand.Seed(time.Now().Unix())
	parseArgs()
	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	var wait sync.WaitGroup
	wait.Add(NR_CLIENTS)
	for i := 0; i < NR_CLIENTS; i++ {
		bot := newClientBot(i+1, &wait)
		go bot.run()
	}
	timer.StartTicks(time.Millisecond * 100)
	wait.Wait()
}
