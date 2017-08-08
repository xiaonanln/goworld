package kvdb

import (
	"testing"

	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/kvdb/types"
)

func init() {
	config.SetConfigFile("../../goworld.ini")
	Initialize()
}

func TestBasic(t *testing.T) {
	Get("__key_not_exists__", func(val string, err error) {
		if err != nil {
			t.Error(err)
			return
		}
		if val != "" {
			t.Fail()
		}
	})

	Put("a", "111", func(err error) {
		if err != nil {
			t.Error(err)
			return
		}
		Get("a", func(val string, err error) {
			if err != nil {
				t.Error(err)
				return
			}
			if val != "111" {
				t.Fail()
			}
		})
	})

	GetRange("a", "z", func(items []kvdbtypes.KVItem, err error) {
		if err != nil {
			t.Error(err)
			return
		}
	})
}

func TestClose(t *testing.T) {
	Close()
	Initialize()
}
