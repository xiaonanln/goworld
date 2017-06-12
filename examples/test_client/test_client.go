package main

import (
	"flag"

	"github.com/xiaonanln/goworld/config"
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


}
