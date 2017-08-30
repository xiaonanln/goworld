package entitystorageredis

import (
	"database/sql"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/storage/storage_common"

	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dataPacker = netutil.MessagePackMsgPacker{}
)

type mysqlEntityStorage struct {
	db                 *sql.DB
	visitedEntityTypes common.StringSet
}

// OpenMySQL opens redis as entity storage
func OpenMySQL(url string) (storagecommon.EntityStorage, error) {
	db, err := sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &mysqlEntityStorage{
		db:                 db,
		visitedEntityTypes: common.StringSet{},
	}, nil
}

func (es *mysqlEntityStorage) createTableForEntityTypeIfNotExists(typeName string) error {
	if es.visitedEntityTypes.Contains(typeName) {
		return nil
	}

	es.visitedEntityTypes.Add(typeName)
	_, err := es.db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s`(`id` CHAR(%d) NOT NULL PRIMARY KEY, `data` BLOB NOT NULL)", typeName, common.ENTITYID_LENGTH))
	if err != nil {
		return err
	}
	return nil
}

func escapeId(id string) string {
	return "`" + id + "`"
}

func packData(data interface{}) (b []byte, err error) {
	b, err = dataPacker.PackMsg(data, b)
	return
}

func (es *mysqlEntityStorage) List(typeName string) ([]common.EntityID, error) {
	es.createTableForEntityTypeIfNotExists(typeName)

	return nil, nil
}

func (es *mysqlEntityStorage) Write(typeName string, entityID common.EntityID, data interface{}) error {
	es.createTableForEntityTypeIfNotExists(typeName)

	_, err := packData(data)
	if err != nil {
		return err
	}
	return nil
}

func (es *mysqlEntityStorage) Read(typeName string, entityID common.EntityID) (interface{}, error) {
	return nil, nil
	//var err error
	//var data map[string]interface{}
	//if err = dataPacker.UnpackMsg(b, &data); err != nil {
	//	return nil, err
	//}
	//return data, nil
}

func (es *mysqlEntityStorage) Exists(typeName string, entityID common.EntityID) (bool, error) {
	return false, nil
}

func (es *mysqlEntityStorage) Close() {
	es.db.Close()
}

func (es *mysqlEntityStorage) IsEOF(err error) bool {
	return true
}
