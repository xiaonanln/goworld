package goworld

import (
	"github.com/xiaonanln/goworld/components/game"
	"github.com/xiaonanln/goworld/entity"
)

func Run(delegate game.IGameDelegate) {
	game.Run(delegate)
}

func RegisterEntity(typeName string, entityPtr entity.IEntity) {
	entity.RegisterEntity(typeName, entityPtr)
}

func CreateEntity(typeName string) {
	game.CreateEntity(typeName)
}
