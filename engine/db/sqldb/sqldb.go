package sqldb

import (
	"database/sql"

	"github.com/xiaonanln/goworld/engine/async"
	"golang.org/x/net/context"
)

const (
	_SQLDB_ASYNC_JOB_GROUP = "_sqldb"
)

type DB struct {
	db *sql.DB
}

type Tx struct {
	*sql.Tx
}

func Open(driverName, dataSourceName string, ac async.AsyncCallback) {
	async.AppendAsyncJob(_SQLDB_ASYNC_JOB_GROUP, func() (res interface{}, err error) {
		db, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}

		return &DB{db}, err
	}, ac)
}

func (db *DB) Close(ac async.AsyncCallback) {
	async.AppendAsyncJob(_SQLDB_ASYNC_JOB_GROUP, func() (res interface{}, err error) {
		err = db.db.Close()
		return
	}, ac)
}

func (db *DB) Ping(ac async.AsyncCallback) {
	async.AppendAsyncJob(_SQLDB_ASYNC_JOB_GROUP, func() (res interface{}, err error) {
		err = db.db.Ping()
		return
	}, ac)
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions, ac async.AsyncCallback) {
	async.AppendAsyncJob(_SQLDB_ASYNC_JOB_GROUP, func() (res interface{}, err error) {
		tx, err := db.db.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		} else {
			return &Tx{tx}, err
		}
	}, ac)
}
