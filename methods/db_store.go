package methods

import (
	"errors"
	"fmt"
	"goseed/log"
	"goseed/schemas"

	"github.com/jmoiron/sqlx"
)

func NewDbStore(db *sqlx.DB) *DbStore {
	return &DbStore{DB: db}
}

type DbStore struct {
	*sqlx.DB
}

func (s *DbStore) Setup() error {
	fmt.Println("setup database")
	_, err := s.Exec("CREATE DATABASE IF NOT EXISTS goseed;")
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	return nil
}

func (s *DbStore) UseDatabase(name string) error {
	_, err := s.Exec("USE " + name + ";")
	if err != nil {
		return fmt.Errorf("failed to use database: %w", err)
	}
	log.Success("database is now '" + name + "'")
	return nil
}

func (s *DbStore) GetTableFields(database, table string) ([]schemas.TableFields, error) {
	result := []schemas.TableFields{}
	s.Select(&result, fmt.Sprintf("SELECT COLUMN_NAME AS 'Field', COLUMN_TYPE AS `Type`, IS_NULLABLE AS `NULL`, COLUMN_KEY AS `Key`,COLUMN_DEFAULT AS `Default`, EXTRA AS `Extra` FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s';", database, table))
	return result, nil
}

func (s *DbStore) GenerateInsertionMap([]schemas.TableFields) map[string]interface{} {
	var result map[string]interface{}
	return result
}

func GenerateTableFieldValue(fields schemas.TableFields, index int) (string, NumberNil, error) {
	if fields.Extra != nil {
		if len(*fields.Extra) > 0 {
			if *fields.Extra == "auto_increment" {
				return "auto_increment", NumberImpl{number: index + 1}, nil
			}
		}
	}
	return "", nil, errors.New("failed to generate table field value")
}

type NumberNil interface {
	Number() int
}

type NumberImpl struct {
	number int
}

func (n NumberImpl) Number() int {
	return n.number
}
