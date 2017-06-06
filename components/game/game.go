package game

import "flag"

var (
	gameid int
)

func init() {
	parseArgs()
}

func parseArgs() {
	flag.IntVar(&gameid, "gid", 0, "set gameid")
	flag.Parse()
}

func Run(delegate IGameDelegate) {
	newGameService(gameid, delegate).run()
}
