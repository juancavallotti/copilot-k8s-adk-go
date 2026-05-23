package traces

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	types "juancavallotti.com/recipe-types"
)

var (
	ErrEmptyEventID   = errors.New("service: empty event_id")
	ErrZeroOccurredAt = errors.New("service: zero occurred_at")
	ErrEmptyTraceData = errors.New("service: empty trace data")
)

func (s *Service) LogTrace(ctx context.Context, eventID string, occurredAt time.Time, data json.RawMessage) error {
	if strings.TrimSpace(eventID) == "" {
		return ErrEmptyEventID
	}
	if occurredAt.IsZero() {
		return ErrZeroOccurredAt
	}
	if len(data) == 0 {
		return ErrEmptyTraceData
	}
	return s.store.InsertTrace(ctx, eventID, occurredAt, data)
}

func (s *Service) ListEvents(ctx context.Context, limit, offset int) ([]types.Event, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.store.ListEvents(ctx, limit, offset)
}

func (s *Service) DeleteAllEvents(ctx context.Context) error {
	return s.store.DeleteAllEvents(ctx)
}

func (s *Service) DeleteEvent(ctx context.Context, eventID string) error {
	if strings.TrimSpace(eventID) == "" {
		return ErrEmptyEventID
	}
	return s.store.DeleteEventByID(ctx, eventID)
}

func (s *Service) ListTracesByEvent(ctx context.Context, eventID string, limit, offset int) ([]types.Trace, error) {
	if strings.TrimSpace(eventID) == "" {
		return nil, ErrEmptyEventID
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.store.ListTracesByEvent(ctx, eventID, limit, offset)
}
