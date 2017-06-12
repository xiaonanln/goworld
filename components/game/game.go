package game

import (
	"flag"

	"math/rand"
	"time"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/config"
)

var (
	gameid      int
	configFile  string
	gameService *GameService
)

func init() {
	parseArgs()
}

func parseArgs() {
	flag.IntVar(&gameid, "gid", 0, "set gameid")
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.Parse()
}

func Run(delegate IGameDelegate) {
	rand.Seed(time.Now().Unix())

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	gameService = newGameService(gameid, delegate)
	gameService.run()
}

func GetServiceProviders(serviceName string) []common.EntityID {
	return gameService.registeredServices[serviceName].ToList()
}
