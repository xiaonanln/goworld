package kvdb_types

// Interface for iterators for KVDB
//
// Next should returns the next item with error=nil whenever has next item
// otherwise returns KVItem{}, io.EOF
// When failed, returns KVItem{}, error
type Iterator interface {
	Next() (KVItem, error)
}

type KVItem struct {
	Key string
	Val string
}
