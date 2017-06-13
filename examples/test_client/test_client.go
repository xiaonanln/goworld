package main

import (
	"flag"

	"sync"

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
	parseArgs()
	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	var wait sync.WaitGroup
	wait.Add(NR_CLIENTS)
	for i := 0; i < NR_CLIENTS; i++ {
		bot := newClientBot(&wait)
		go bot.run()
	}
	wait.Wait()
}
