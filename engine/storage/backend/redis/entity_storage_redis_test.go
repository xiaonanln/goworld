package entitystorageredis

import (
	"testing"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/typeconv"
)

func TestRedisEntityStorage(t *testing.T) {
	es, err := OpenRedis("redis://localhost:6379", 0)
	if err != nil {
		t.Error(err)
	}
	gwlog.Infof("TestRedisEntityStorage: %v", es)
	entityID := common.GenEntityID()
	gwlog.Infof("TESTING ENTITYID: %s", entityID)
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

	if typeconv.Int(verifyData.(map[string]interface{})["a"]) != 1 {
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

	gwlog.Infof("Found avatars saved: %v", avatarIDs)
	for _, avatarID := range avatarIDs {
		data, err := es.Read("Avatar", avatarID)
		if err != nil {
			t.Error(err)
		}
		t.Logf("Read Avatar %s => %v", avatarID, data)
	}

}
