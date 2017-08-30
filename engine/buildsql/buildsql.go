package buildsql

type SQLBuilder interface {
	CreateTable(tableName string, ifNotExists bool) (sql string, args []interface{})
}
