package main

import (
	"flag"

	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/components/game"
	"github.com/xiaonanln/goworld/entity"
)

var (
	gameid = 0
)

type TestEntity struct {
	entity.Entity
}

func init() {

}

type gameDelegate struct {
	game.GameDelegate
}

func main() {
	parseArgs()

	goworld.RegisterEntity("TestEntity", &TestEntity{})
	goworld.Run(gameid, &gameDelegate{})
}

func parseArgs() {
	flag.IntVar(&gameid, "gid", 0, "set gameid")
	flag.Parse()
}

func (game gameDelegate) OnReady() {
	game.GameDelegate.OnReady()
	goworld.CreateEntity("TestEntity")
}
