package schemas

type DbStore interface {
	Setup() error
	UseDatabase(name string) error
	GetTableFields(database, table string) ([]TableFields, error)
	GenerateInsertionMap(fields []TableFields, seedSize int64) []map[string]InsertionMap
	BatchInsertFromMap(bArr []map[string]InsertionMap, fields []TableFields, table string, chunkSize int64) error
	SelectCount(table string) (int64, error)
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
