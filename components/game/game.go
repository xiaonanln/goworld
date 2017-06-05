package game

import (
	"fmt"

	"os"

	"time"

	"flag"

	"github.com/xiaonanln/goTimer"
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

	timer.AddCallback(0, func() {
		gameDelegate.OnReady()
	})

	for {
		timer.Tick()
		os.Stderr.Write([]byte{'.'})
		time.Sleep(time.Millisecond * 100)
	}
}
