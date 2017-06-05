package goworld

import (
	"github.com/xiaonanln/goworld/components/game"
	"github.com/xiaonanln/goworld/entity"
)

func Run(gameid int, delegate game.IGameDelegate) {
	game.Run(gameid, delegate)
}

func RegisterEntity(typeName string, entityPtr entity.IEntity) {
	entity.RegisterEntity(typeName, entityPtr)
}

func CreateEntity(typeName string) {
	entity.CreateEntity(typeName)
}
