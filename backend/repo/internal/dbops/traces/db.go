package traces

import (
	"database/sql"
	"errors"
	"sync"

	"juancavallotti.com/recipes-repo/internal/embeddings"
)

// Store runs trace persistence against a *sql.DB connection pool.
// Like the recipes store, it owns an embedding client (used by the
// async hook fired on InsertTrace) and a WaitGroup so callers can
// drain in-flight indexing before closing the pool.
type Store struct {
	db    *sql.DB
	embed embeddings.Client
	wg    sync.WaitGroup
}

var errNilDB = errors.New("dbops/traces: nil *sql.DB")

// StoreOption configures a Store at construction time.
type StoreOption func(*Store)

// WithEmbedClient overrides the default no-op embedding client.
func WithEmbedClient(c embeddings.Client) StoreOption {
	return func(s *Store) {
		if c != nil {
			s.embed = c
		}
	}
}

// NewStore returns a Store that uses pool for all queries. By default
// the embedding client is a no-op, so existing tests don't have to
// thread one through.
func NewStore(pool *sql.DB, opts ...StoreOption) *Store {
	s := &Store{db: pool, embed: embeddings.Noop{}}
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
