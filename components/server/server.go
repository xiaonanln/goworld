package server

import (
	"flag"

	"math/rand"
	"time"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
)

var (
	serverid    int
	configFile  string
	gameService *GameService
	gateService *GateService
)

func init() {
	parseArgs()
}

func parseArgs() {
	flag.IntVar(&serverid, "sid", 0, "set serverid")
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.Parse()
}

func Run(delegate IServerDelegate) {
	rand.Seed(time.Now().Unix())

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	gateService = newGateService()
	go gateService.run() // run gate service in another goroutine

	gameService = newGameService(serverid, delegate)
	gameService.run()
}

func GetServiceProviders(serviceName string) []common.EntityID {
	return gameService.registeredServices[serviceName].ToList()
}
