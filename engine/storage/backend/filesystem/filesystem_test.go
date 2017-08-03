package entity_storage_filesystem

import (
	"testing"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

func TestFileSystemEntityStorage(t *testing.T) {
	es, err := OpenDirectory("test_entity_storage")
	if err != nil {
		t.Error(err)
	}
	gwlog.Info("TestOpenDirectory: %v", es)
	entityID := common.GenEntityID()
	gwlog.Info("TESTING ENTITYID: %s", entityID)
	data, err := es.Read("Avatar", entityID)
	if data != nil {
		t.Errorf("should be nil")
	}

	testData := map[string]interface{}{
		"a": 1,
		"b": "2",
		"c": true,
		"d": 1.11,
	}
	es.Write("Avatar", entityID, testData)

	verifyData, err := es.Read("Avatar", entityID)
	if err != nil {
		t.Error(err)
	}

	if verifyData.(map[string]interface{})["a"].(float64) != 1 {
		t.Errorf("read wrong data: %v", verifyData)
	}
	if verifyData.(map[string]interface{})["b"].(string) != "2" {
		t.Errorf("read wrong data: %v", verifyData)
	}
	if verifyData.(map[string]interface{})["c"].(bool) != true {
		t.Errorf("read wrong data: %v", verifyData)
	}
	if verifyData.(map[string]interface{})["d"].(float64) != 1.11 {
		t.Errorf("read wrong data: %v", verifyData)
	}

	avatarIDs, err := es.List("Avatar")
	if err != nil {
		t.Error(err)
	}
	if len(avatarIDs) == 0 {
		t.Errorf("Avatar IDs is empty!")
	}

	gwlog.Info("Found avatars saved: %v", avatarIDs)
	for _, avatarID := range avatarIDs {
		data, err := es.Read("Avatar", avatarID)
		if err != nil {
			t.Error(err)
		}
		t.Logf("Read Avatar %s => %v", avatarID, data)
	}

}
