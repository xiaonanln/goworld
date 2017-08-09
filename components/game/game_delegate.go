package game

import "github.com/xiaonanln/goworld/engine/gwlog"

// IGameDelegate defines interfaces for handling game server events
type IGameDelegate interface {
	OnGameReady()
}

// GameDelegate is the default IGameDelegate implementation
type GameDelegate struct {
}

// OnGameReady is called when game is ready
func (gd *GameDelegate) OnGameReady() {
	gwlog.Infof("game %d is ready.", gameid)
}
