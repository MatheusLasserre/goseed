package schemas

type Person struct {
	ID        int    `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	CreatedAt string `json:"created_at" db:"created_at"`
}

type PersonStore interface {
	GetAll() ([]Person, error)
	Setup() error
	Seed() error
}
