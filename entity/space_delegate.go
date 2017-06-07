package entity

import "github.com/xiaonanln/goworld/gwlog"

var (
	spaceDelegate ISpaceDelegate = &DefaultSpaceDelegate{}
)

func SetSpaceDelegate(delegate ISpaceDelegate) {
	spaceDelegate = delegate
}

// Space delegate interface
type ISpaceDelegate interface {
	OnSpaceCreated(space *Space)
}

// The default space delegate
type DefaultSpaceDelegate struct {
}

func (delegate *DefaultSpaceDelegate) OnSpaceCreated(space *Space) {
	gwlog.Debug("OnSpaceCreated: %s", space)
}
