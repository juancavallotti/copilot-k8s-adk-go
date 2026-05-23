package traces

import (
	"context"
	"encoding/json"
	"time"

	types "juancavallotti.com/recipe-types"
)

type store interface {
	InsertTrace(ctx context.Context, eventID string, occurredAt time.Time, data json.RawMessage) error
	ListEvents(ctx context.Context, limit, offset int) ([]types.Event, error)
	ListTracesByEvent(ctx context.Context, eventID string, limit, offset int) ([]types.Trace, error)
	DeleteAllEvents(ctx context.Context) error
	DeleteEventByID(ctx context.Context, eventID string) error
}

type Service struct {
	store store
}

// NewService wires a trace store into the trace service layer.
func NewService(store store) *Service {
	return &Service{store: store}
}
