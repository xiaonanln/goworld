package entity

type ISpace interface {
	OnSpaceInit()    // Called when initializing space struct, override to initialize custom space fields
	OnSpaceCreated() // Called when space is created
	OnSpaceDestroy() // Called just before space is destroyed
	// Space Operations
	OnEntityEnterSpace(entity *Entity) // Called when any entity enters space
	OnEntityLeaveSpace(entity *Entity) // Called when any entity leaves space
}
