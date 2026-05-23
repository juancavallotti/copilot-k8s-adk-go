package skills

import (
	"database/sql"
	"errors"
)

// Store runs skill persistence against a *sql.DB connection pool.
type Store struct {
	db *sql.DB
}

var errNilDB = errors.New("dbops/skills: nil *sql.DB")

// NewStore returns a Store that uses pool for all queries.
func NewStore(pool *sql.DB) *Store {
	return &Store{db: pool}
}
