package entity

type ISpace interface {
	OnSpaceInit()
	OnSpaceCreated()
	OnSpaceDestroy()
	// Space Operations
	OnEntityEnterSpace(entity *Entity)
	OnEntityLeaveSpace(entity *Entity)
}
