package server

import (
	"github.com/google/btree"
	"github.com/xiaonanln/goworld/common"
)

const (
	FILTER_TREE_DEGREE = 2
)

type FilterTree struct {
	btree *btree.BTree
}

func NewFilterTree() *FilterTree {
	return &FilterTree{
		btree: btree.New(FILTER_TREE_DEGREE),
	}
}

//func (ft *FilterTree) Insert(clientid string) {
//
//}

type FilterTreeItem struct {
	clientid common.ClientID
	val      string
}

func (it *FilterTreeItem) Less(_other btree.Item) bool {
	other := _other.(*FilterTreeItem)
	return it.val < other.val || (it.val == other.val && it.clientid < other.clientid)
}

func (ft *FilterTree) Insert(id common.ClientID, val string) {
	ft.btree.ReplaceOrInsert(&FilterTreeItem{
		clientid: id,
		val:      val,
	})
}

func (ft *FilterTree) Remove(id common.ClientID, val string) {
	//gwlog.Info("Removing %s %s has %v", id, val, ft.btree.Has(&FilterTreeItem{
	//	clientid: id,
	//	val:      val,
	//}))

	ft.btree.Delete(&FilterTreeItem{
		clientid: id,
		val:      val,
	})
}

func (ft *FilterTree) Visit(val string, f func(clientid common.ClientID)) {
	ft.btree.AscendGreaterOrEqual(&FilterTreeItem{common.ClientID(""), val}, func(_item btree.Item) bool {
		item := _item.(*FilterTreeItem)
		if item.val > val {
			return false
		}

		f(item.clientid)
		return true
	})
}
