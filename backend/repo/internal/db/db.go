package db

import (
	"context"
	"database/sql"
	"errors"

	types "juancavallotti.com/recipe-types"
)

// DB is the persistence contract for recipes.
type DB interface {
	GetRecipes(ctx context.Context) ([]types.Recipe, error)
	GetRecipe(ctx context.Context, id string) (types.Recipe, error)
	CreateRecipe(ctx context.Context, recipe types.Recipe) error
	UpdateRecipe(ctx context.Context, recipe types.Recipe) error
	DeleteRecipe(ctx context.Context, id string) error
}

// Store implements DB using a *sql.DB connection pool.
type Store struct {
	db *sql.DB
}

var errNilDB = errors.New("db: nil *sql.DB")

// NewStore returns a Store that uses pool for all queries.
func NewStore(pool *sql.DB) *Store {
	return &Store{db: pool}
}

// Compile-time assertion that Store implements DB.
var _ DB = (*Store)(nil)
