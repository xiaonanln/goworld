package main

import (
	"unsafe"

	"github.com/petar/GoLLRB/llrb"
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
	cp  *ClientProxy
	val string
}

func (it *filterTreeItem) Less(_other llrb.Item) bool {
	other := _other.(*filterTreeItem)
	return it.val < other.val || (it.val == other.val && uintptr(unsafe.Pointer(it.cp)) < uintptr(unsafe.Pointer(other.cp)))
}

func (ft *_FilterTree) Insert(cp *ClientProxy, val string) {
	ft.btree.ReplaceOrInsert(&filterTreeItem{
		cp:  cp,
		val: val,
	})
}

func (ft *_FilterTree) Remove(cp *ClientProxy, val string) {
	//gwlog.Infof("Removing %s %s has %v", id, val, ft.llrb.Has(&filterTreeItem{
	//	cp: id,
	//	val:      val,
	//}))

	ft.btree.Delete(&filterTreeItem{
		cp:  cp,
		val: val,
	})
}

func (ft *_FilterTree) Visit(op proto.FilterClientsOpType, val string, f func(cp *ClientProxy)) {
	if op == proto.FILTER_CLIENTS_OP_EQ {
		// visit key == val
		ft.btree.AscendGreaterOrEqual(&filterTreeItem{nil, val}, func(_item llrb.Item) bool {
			item := _item.(*filterTreeItem)
			if item.val > val {
				return false
			}

			f(item.cp)
			return true
		})
	} else if op == proto.FILTER_CLIENTS_OP_NE {
		// visit key != val
		// visit key < val first
		ft.btree.AscendLessThan(&filterTreeItem{nil, val}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).cp)
			return true
		})
		// then visit key > val
		ft.btree.AscendGreaterOrEqual(&filterTreeItem{nil, gwutils.NextLargerKey(val)}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).cp)
			return true
		})
	} else if op == proto.FILTER_CLIENTS_OP_GT {
		// visit key > val
		ft.btree.AscendGreaterOrEqual(&filterTreeItem{nil, gwutils.NextLargerKey(val)}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).cp)
			return true
		})
	} else if op == proto.FILTER_CLIENTS_OP_GTE {
		// visit key >= val
		ft.btree.AscendGreaterOrEqual(&filterTreeItem{nil, val}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).cp)
			return true
		})
	} else if op == proto.FILTER_CLIENTS_OP_LT {
		// visit key < val
		ft.btree.AscendLessThan(&filterTreeItem{nil, val}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).cp)
			return true
		})
	} else if op == proto.FILTER_CLIENTS_OP_LTE {
		// visit key <= val
		ft.btree.AscendLessThan(&filterTreeItem{nil, gwutils.NextLargerKey(val)}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).cp)
			return true
		})
	} else {
		gwlog.Panicf("unknown filter clients op: %s", op)
	}
}
