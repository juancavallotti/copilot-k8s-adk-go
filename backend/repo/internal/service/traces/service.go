package traces

import (
	"context"
	"encoding/json"
	"time"

	types "juancavallotti.com/recipe-types"
	traceops "juancavallotti.com/recipes-repo/internal/dbops/traces"
)

type store interface {
	InsertTrace(ctx context.Context, eventID string, occurredAt time.Time, data json.RawMessage) error
	ListEvents(ctx context.Context, limit, offset int) ([]types.Event, error)
	ListTracesByEvent(ctx context.Context, eventID string, limit, offset int) ([]types.Trace, error)
	DeleteAllEvents(ctx context.Context) error
	DeleteEventByID(ctx context.Context, eventID string) error
	IndexEvent(ctx context.Context, eventID string, force bool) error
	ReindexEvents(ctx context.Context, opts traceops.ReindexEventsOptions) error
	SearchEvents(ctx context.Context, query string, limit int) ([]types.EventMatch, error)
	Wait()
}

type Service struct {
	store store
}

// NewService wires a trace store into the trace service layer.
func NewService(store store) *Service {
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
