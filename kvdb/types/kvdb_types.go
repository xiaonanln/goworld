package kvdb_types

type Iterator interface {
	Next() (KVItem, error)
}

type KVItem struct {
	Key string
	Val string
}
