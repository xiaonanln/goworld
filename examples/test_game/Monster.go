package main

import "github.com/xiaonanln/goworld/engine/entity"

// Monster type
type Monster struct {
	entity.Entity // Entity type should always inherit entity.Entity
}

func (m *Monster) DefineAttrs(desc *entity.EntityTypeDesc) {
	desc.DefineAttr("name", "AllClients")
}
