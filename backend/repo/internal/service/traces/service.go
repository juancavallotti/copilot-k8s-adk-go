package traces

import (
	"context"

	types "juancavallotti.com/recipe-types"
	traceops "juancavallotti.com/recipes-repo/internal/dbops/traces"
)

type Service struct {
	store *traceops.Store
}

// NewService wires a trace store into the trace service layer.
func NewService(store *traceops.Store) *Service {
	return &Service{store: store}
}

// IndexEvent rebuilds the embedding row for a single event.
func (s *Service) IndexEvent(ctx context.Context, eventID string, force bool) error {
	return s.store.IndexEvent(ctx, eventID, force)
}

// ReindexEvents streams a bulk reindex pass through the store.
func (s *Service) ReindexEvents(ctx context.Context, opts traceops.ReindexEventsOptions) error {
	return s.store.ReindexEvents(ctx, opts)
}

// SearchEvents runs a semantic search and returns ranked events.
func (s *Service) SearchEvents(ctx context.Context, query string, limit int) ([]types.EventMatch, error) {
	return s.store.SearchEvents(ctx, query, limit)
}

// Wait blocks until in-flight async event-embedding work in the
// store completes.
func (s *Service) Wait() {
	s.store.Wait()
}
