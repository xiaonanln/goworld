package entity

import (
	"testing"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/netutil"
)

type TestEntity struct {
	Entity
}

func (e *TestEntity) DescribeEntityType(*EntityTypeDesc) {

}

func TestMigrateData(t *testing.T) {
	RegisterEntity("TestEntity", &TestEntity{}, false)
	e := CreateEntityLocally("TestEntity", nil)
	t.Logf("created entity %s", e)
	targetSpaceID := common.GenEntityID()
	e.Attrs.SetBool("bool", true)
	e.Attrs.SetStr("str", "strval")
	md := e.GetMigrateData(targetSpaceID, e.Position)
	t.Logf("get migrate data: %+v", md)
	data, err := netutil.MSG_PACKER.PackMsg(md, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("pack migrate data: %d bytes", len(data))

	var umd entityMigrateData
	if err := netutil.MSG_PACKER.UnpackMsg(data, &umd); err != nil {
		t.Fatal(err)
	}

	t.Logf("Pack Migrate Data: %+v", md)
	t.Logf("UnPack Migrate Data: %+v", umd)
	if umd.SpaceID != md.SpaceID {
		t.Fatalf("SpaceID mismatch: %#v & %#v", umd.SpaceID, md.SpaceID)
	}
	if sv, ok := umd.Attrs["str"]; !ok || sv.(string) != "strval" {
		t.Fatalf("str is not strval")
	}
	if bv, ok := umd.Attrs["bool"]; !ok || bv.(bool) != true {
		t.Fatalf("bool is not true")
	}
}
