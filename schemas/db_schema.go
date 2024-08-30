package schemas

import "sync"

type DbStore interface {
	Setup(relFilePath string) error
	GetTableFields(database, table string) ([]TableFields, error)
	GenerateInsertionMap(fields []TableFields, table string, seedSize int64, chunkSize int64, maxConn int, dbName string, wg *sync.WaitGroup) error
	BatchInsertFromMap(bArr []map[string]InsertionMap, fields []TableFields, table string, chunkSize int64, dbName string, maxConn int) error
	SelectCount(table string, dbName string) (int64, error)
	GetMaxConnections() (int, error)
}

type TableFields struct {
	Field   string  `db:"Field"`
	Type    string  `db:"Type"`
	Null    string  `db:"NULL"`
	Key     *string `db:"Key"`
	Default *string `db:"Default"`
	Extra   *string `db:"Extra"`
}

type InsertionMap struct {
	StrValue string
	IntValue NumberNil
}

type NumberNil interface {
	Number() int64
}
type ShowVariables struct {
	Variable_name string `db:"Variable_name"`
	Value         string `db:"Value"`
}
