package main

import (
	llrb "github.com/petar/GoLLRB/llrb"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/proto"
)

type _FilterTree struct {
	btree *llrb.LLRB
}

func newFilterTree() *_FilterTree {
	return &_FilterTree{
		btree: llrb.New(),
	}
}

type filterTreeItem struct {
	clientid common.ClientID
	val      string
}

func (it *filterTreeItem) Less(_other llrb.Item) bool {
	other := _other.(*filterTreeItem)
	return it.val < other.val || (it.val == other.val && it.clientid < other.clientid)
}

func (ft *_FilterTree) Insert(id common.ClientID, val string) {
	ft.btree.ReplaceOrInsert(&filterTreeItem{
		clientid: id,
		val:      val,
	})
}

func (ft *_FilterTree) Remove(id common.ClientID, val string) {
	//gwlog.Infof("Removing %s %s has %v", id, val, ft.llrb.Has(&filterTreeItem{
	//	clientid: id,
	//	val:      val,
	//}))

	ft.btree.Delete(&filterTreeItem{
		clientid: id,
		val:      val,
	})
}

func (ft *_FilterTree) Visit(op proto.FilterClientsOpType, val string, f func(clientid common.ClientID)) {
	if op == proto.FILTER_CLIENTS_OP_EQ {
		// visit key == val
		ft.btree.AscendGreaterOrEqual(&filterTreeItem{common.ClientID(""), val}, func(_item llrb.Item) bool {
			item := _item.(*filterTreeItem)
			if item.val > val {
				return false
			}

			f(item.clientid)
			return true
		})
	} else if op == proto.FILTER_CLIENTS_OP_NE {
		// visit key != val
		// visit key < val first
		ft.btree.AscendLessThan(&filterTreeItem{common.ClientID(""), val}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).clientid)
			return true
		})
		// then visit key > val
		ft.btree.AscendGreaterOrEqual(&filterTreeItem{common.ClientID(""), gwutils.NextLargerKey(val)}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).clientid)
			return true
		})
	} else if op == proto.FILTER_CLIENTS_OP_GT {
		// visit key > val
		ft.btree.AscendGreaterOrEqual(&filterTreeItem{common.ClientID(""), gwutils.NextLargerKey(val)}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).clientid)
			return true
		})
	} else if op == proto.FILTER_CLIENTS_OP_GTE {
		// visit key >= val
		ft.btree.AscendGreaterOrEqual(&filterTreeItem{common.ClientID(""), val}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).clientid)
			return true
		})
	} else if op == proto.FILTER_CLIENTS_OP_LT {
		// visit key < val
		ft.btree.AscendLessThan(&filterTreeItem{common.ClientID(""), val}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).clientid)
			return true
		})
	} else if op == proto.FILTER_CLIENTS_OP_LTE {
		// visit key <= val
		ft.btree.AscendLessThan(&filterTreeItem{common.ClientID(""), gwutils.NextLargerKey(val)}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).clientid)
			return true
		})
	} else {
		gwlog.Panicf("unknown filter clients op: %s", op)
	}
}
