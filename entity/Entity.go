package entity

import "github.com/xiaonanln/goworld/uuid"

type EntityID string

func GenEntityID() EntityID {
	return EntityID(uuid.GenUUID())
}

type Entity struct {
	ID EntityID
}

type IEntity interface {
}
