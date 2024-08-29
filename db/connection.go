package db

import (
	"fmt"
	"goseed/log"
	"goseed/schemas"
	"goseed/store"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func Connect(conn string) (*Store, error) {
	db, err := sqlx.Open("mysql", conn)
	if err != nil {
		return nil, fmt.Errorf("failed to opening to database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Success("Database connected.")
	return &Store{
		DbStore: store.NewDbStore(db),
		DB:      db,
	}, nil
}

type Store struct {
	schemas.DbStore
	*sqlx.DB
}

// root:goseed@tcp(localhost:3306)/
