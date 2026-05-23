package recipes

import (
	"database/sql"
	"errors"
)

// Store runs recipe persistence against a *sql.DB connection pool.
type Store struct {
	db *sql.DB
}

var errNilDB = errors.New("dbops/recipes: nil *sql.DB")

// NewStore returns a Store that uses pool for all queries.
func NewStore(pool *sql.DB) *Store {
	return &Store{db: pool}
}
