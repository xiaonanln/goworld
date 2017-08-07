package main

import (
	"github.com/google/btree"
	"github.com/xiaonanln/goworld/engine/common"
)

const (
	FILTER_TREE_DEGREE = 2
)

type _FilterTree struct {
	btree *btree.BTree
}

func newFilterTree() *_FilterTree {
	return &_FilterTree{
		btree: btree.New(FILTER_TREE_DEGREE),
	}
}

type filterTreeItem struct {
	clientid common.ClientID
	val      string
}

func (it *filterTreeItem) Less(_other btree.Item) bool {
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
	//gwlog.Info("Removing %s %s has %v", id, val, ft.btree.Has(&filterTreeItem{
	//	clientid: id,
	//	val:      val,
	//}))

	ft.btree.Delete(&filterTreeItem{
		clientid: id,
		val:      val,
	})
}

func (ft *_FilterTree) Visit(val string, f func(clientid common.ClientID)) {
	ft.btree.AscendGreaterOrEqual(&filterTreeItem{common.ClientID(""), val}, func(_item btree.Item) bool {
		item := _item.(*filterTreeItem)
		if item.val > val {
			return false
		}

		f(item.clientid)
		return true
	})
}
