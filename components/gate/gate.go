package main

import (
	"flag"

	"github.com/xiaonanln/goworld/config"
)

var (
	gateid int
)

func parseArgs() {
	flag.IntVar(&gateid, "gid", 0, "set gateid")
	flag.Parse()
}

func main() {
	parseArgs()
	gatecfg := config.GetGate(gateid)
	newGateServer(gatecfg).run()
}
