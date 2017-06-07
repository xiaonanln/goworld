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

//func createEntity(typeName string) {
//	game.createEntity(typeName)
//}

func CreateSpace() {
	entity.CreateSpace()
}

func SetSpaceDelegate(delegate entity.ISpaceDelegate) {
	entity.SetSpaceDelegate(delegate)
}
