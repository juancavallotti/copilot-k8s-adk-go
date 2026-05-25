package recipes

import (
	"database/sql"
	"errors"

	"juancavallotti.com/recipes-repo/internal/embeddings"
)

// Store runs recipe persistence against a *sql.DB connection pool.
// It also owns the embedding client used by write hooks and the
// reindex command.
type Store struct {
	db    *sql.DB
	embed embeddings.Client
}

var errNilDB = errors.New("dbops/recipes: nil *sql.DB")

// StoreOption configures a Store at construction time.
type StoreOption func(*Store)

// WithEmbedClient overrides the default no-op embedding client. The
// Repo constructor calls this; tests can leave it unset and the store
// will silently skip async indexing.
func WithEmbedClient(c embeddings.Client) StoreOption {
	return func(s *Store) {
		if c != nil {
			s.embed = c
		}
	}
}

// NewStore returns a Store that uses pool for all queries. By default
// the embedding client is a no-op so callers that don't care about
// indexing (e.g. unit tests with sqlmock) don't have to wire one in.
func NewStore(pool *sql.DB, opts ...StoreOption) *Store {
	s := &Store{db: pool, embed: embeddings.Noop{}}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
