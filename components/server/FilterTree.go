package server

import (
	"sync"

	"github.com/google/btree"
)

const (
	FILTER_TREE_DEGREE = 2
)

type FilterTree struct {
	btree *btree.BTree
	sync.RWMutex
}

func newFilterTree() *FilterTree {
	return &FilterTree{
		btree: btree.New(FILTER_TREE_DEGREE),
	}
}

//func (ft *FilterTree) Insert(id string) {
//
//}

type FilterItem struct {
	id  string
	val string
}

func (it *FilterItem) Less(_other btree.Item) bool {
	other := _other.(*FilterItem)
	if it.val < other.val {
		return true
	}
	if it.id < other.id {
		return true
	}
	return false
}

func (ft *FilterTree) Insert(id string, val string) {
	ft.btree.ReplaceOrInsert(&FilterItem{
		id:  id,
		val: val,
	})
}

func (ft *FilterTree) Remove(id string, val string) {
	ft.btree.Delete(&FilterItem{
		id:  id,
		val: val,
	})
}

func (ft *FilterTree) Visit(val string, f func(string)) {

}
