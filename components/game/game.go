package game

import (
	"fmt"

	"os"

	"time"

	"flag"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/config"
)

var (
	gameid       int
	gameDelegate IGameDelegate
)

func init() {
	parseArgs()
}

func parseArgs() {
	flag.IntVar(&gameid, "gid", 0, "set gameid")
	flag.Parse()
}

func Run(delegate IGameDelegate) {
	gameDelegate = delegate

	cfg := config.GetGame(gameid)
	fmt.Fprintf(os.Stderr, "Read game %d config: \n%s\n", gameid, config.DumpPretty(cfg))

	dispatcher_client.Initialize()

	timer.AddCallback(0, func() {
		gameDelegate.OnReady()
	})

	tickCounter := 0
	for {
		timer.Tick()
		tickCounter += 1
		os.Stderr.Write([]byte{'.'})
		if tickCounter%100 == 0 {
			os.Stderr.Write([]byte{'\n'})
		}

		time.Sleep(time.Millisecond * 100)
	}
}
