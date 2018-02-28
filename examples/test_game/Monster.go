package main

import "github.com/xiaonanln/goworld/engine/entity"

// Monster type
type Monster struct {
	entity.Entity // Entity type should always inherit entity.Entity
}

func (m *Monster) DescribeEntityType(desc *entity.EntityTypeDesc) {
	desc.SetUseAOI(true)
	desc.DefineAttr("name", "AllClients")
}
