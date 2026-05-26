package traces

import (
	"context"
	"database/sql"
	"errors"
	"sync"
)

// Embedder is the slice of embedding-client behavior this package needs.
// Defined here (rather than imported from the embeddings package) so the
// store depends only on what it consumes. A nil Embedder means embeddings
// are disabled — the InsertTrace hook short-circuits and IndexEvent/
// SearchEvents return embeddings.ErrDisabled.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// Store runs trace persistence against a *sql.DB connection pool.
// Like the recipes store, it owns an embedder (used by the async hook
// fired on InsertTrace) and a WaitGroup so callers can drain in-flight
// indexing before closing the pool.
type Store struct {
	db    *sql.DB
	embed Embedder
	wg    sync.WaitGroup
}

var errNilDB = errors.New("dbops/traces: nil *sql.DB")

// StoreOption configures a Store at construction time.
type StoreOption func(*Store)

// WithEmbedClient wires an embedder into the store. Passing nil (or
// omitting the option entirely) leaves embeddings disabled.
func WithEmbedClient(c Embedder) StoreOption {
	return func(s *Store) {
		s.embed = c
	}
}

// NewStore returns a Store that uses pool for all queries. By default
// no embedder is wired in, so existing tests don't have to thread one
// through.
func NewStore(pool *sql.DB, opts ...StoreOption) *Store {
	s := &Store{db: pool}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Wait blocks until all in-flight async event-embedding goroutines
// have completed. The Repo calls this from Close so short-lived CLI
// invocations don't orphan the goroutine fired by InsertTrace.
func (s *Store) Wait() {
	s.wg.Wait()
}
