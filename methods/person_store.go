package methods

import (
	"fmt"
	"goseed/schemas"
	"strconv"

	"github.com/jmoiron/sqlx"
)

func NewPersonStore(db *sqlx.DB) *PersonStore {
	return &PersonStore{DB: db}
}

type PersonStore struct {
	*sqlx.DB
}

func (s *PersonStore) GetAll() ([]schemas.Person, error) {
	var result []schemas.Person
	err := s.Select(&result, "SELECT * FROM person")
	if err != nil {
		return nil, fmt.Errorf("failed to get all person: %w", err)
	}
	return result, nil
}

func (s *PersonStore) Setup() error {
	_, err := s.Exec("DROP TABLE IF EXISTS person;")
	if err != nil {
		return fmt.Errorf("failed to drop person table: %w", err)
	}
	_, err = s.Exec(`CREATE TABLE IF NOT EXISTS person (
	id BIGINT NOT NULL AUTO_INCREMENT, 
	name VARCHAR(255) NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)PRIMARY KEY (id)
	`)
	if err != nil {
		_, err = s.Exec(`CREATE TABLE IF NOT EXISTS person (
			id BIGINT NOT NULL AUTO_INCREMENT, 
			name VARCHAR(255) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id))
			`)
		if err != nil {
			return fmt.Errorf("failed to setup person table: %w", err)
		}
	}
	return nil
}

func (s *PersonStore) Seed() error {
	persons := []schemas.Person{}
	for i := 0; i < 10; i++ {
		persons = append(persons, schemas.Person{
			Name: "Person" + strconv.FormatInt(int64(i), 10),
			ID:   i + 1,
		})
	}
	_, err := s.NamedExec("INSERT INTO person (id, name) VALUES (:id, :name)", persons)
	if err != nil {
		return fmt.Errorf("failed to seed person: %w", err)
	}
	return nil
}
