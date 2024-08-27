package sql

import (
	"fmt"
	"goseed/methods"
	"goseed/schemas"
	"log"

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
	log.Println("connected to database")
	return &Store{
		PersonStore: methods.NewPersonStore(db),
		DbStore:     methods.NewDbStore(db),
		DB:          db,
	}, nil
}

type Store struct {
	schemas.PersonStore
	schemas.DbStore
	*sqlx.DB
}

// root:goseed@tcp(localhost:3306)/
