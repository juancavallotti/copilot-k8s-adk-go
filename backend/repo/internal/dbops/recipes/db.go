package recipes

import (
	"context"
	"database/sql"
	"errors"
	"sync"
)

// Embedder is the slice of embedding-client behavior this package needs.
// Defined here (rather than imported from the embeddings package) so the
// store depends only on what it consumes. A nil Embedder means embeddings
// are disabled — write hooks short-circuit and Search*/Index* return
// embeddings.ErrDisabled.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// Store runs recipe persistence against a *sql.DB connection pool.
// It also owns the embedding client used by write hooks and the
// reindex command, plus a WaitGroup that tracks in-flight async
// indexing goroutines so callers can drain them on shutdown.
type Store struct {
	db    *sql.DB
	embed Embedder
	wg    sync.WaitGroup
}

var errNilDB = errors.New("dbops/recipes: nil *sql.DB")

// StoreOption configures a Store at construction time.
type StoreOption func(*Store)

// WithEmbedClient wires an embedder into the store. Passing nil (or
// omitting the option entirely) leaves embeddings disabled — useful for
// unit tests with sqlmock and for dev environments without an API key.
func WithEmbedClient(c Embedder) StoreOption {
	return func(s *Store) {
		s.embed = c
	}
}

// NewStore returns a Store that uses pool for all queries. By default
// no embedder is wired in, so callers that don't care about indexing
// (e.g. unit tests with sqlmock) don't have to thread one through.
func NewStore(pool *sql.DB, opts ...StoreOption) *Store {
	s := &Store{db: pool}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Wait blocks until all in-flight async embedding writes complete.
// Callers should invoke this before closing the underlying *sql.DB,
// otherwise short-lived processes (CLI invocations) exit before the
// goroutine fired by indexRecipeAsync gets to commit.
func (s *Store) Wait() {
	s.wg.Wait()
}
