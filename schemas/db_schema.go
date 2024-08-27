package schemas

type DbStore interface {
	Setup() error
	UseDatabase(name string) error
	GetTableFields(database, table string) ([]TableFields, error)
}

type TableFields struct {
	Field   string  `db:"Field"`
	Type    string  `db:"Type"`
	Null    string  `db:"NULL"`
	Key     *string `db:"Key"`
	Default *string `db:"Default"`
	Extra   *string `db:"Extra"`
}
