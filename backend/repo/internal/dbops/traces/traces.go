package traces

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	types "juancavallotti.com/recipe-types"
)

// InsertTrace upserts the events row and inserts one trace row in a single
// transaction. data must be valid JSON; occurredAt is the trace's own time
// (taken from the slog record, not the DB clock).
func (s *Store) InsertTrace(ctx context.Context, eventID string, occurredAt time.Time, data json.RawMessage) error {
	if s.db == nil {
		return errNilDB
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	ts := occurredAt.UTC()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO events (event_id, started_at, ended_at, trace_count)
VALUES ($1, $2, $2, 1)
ON CONFLICT (event_id) DO UPDATE SET
    started_at  = LEAST(events.started_at, EXCLUDED.started_at),
    ended_at    = GREATEST(events.ended_at, EXCLUDED.ended_at),
    trace_count = events.trace_count + 1`,
		eventID, ts,
	); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO traces (event_id, occurred_at, data) VALUES ($1, $2, $3::jsonb)`,
		eventID, ts, []byte(data),
	); err != nil {
		return err
	}
	return tx.Commit()
}

// ListEvents returns events newest-first by ended_at. limit caps the page size.
func (s *Store) ListEvents(ctx context.Context, limit, offset int) ([]types.Event, error) {
	if s.db == nil {
		return nil, errNilDB
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT event_id, started_at, ended_at, trace_count
FROM events
ORDER BY ended_at DESC, event_id
LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]types.Event, 0, limit)
	for rows.Next() {
		var e types.Event
		if err := rows.Scan(&e.EventID, &e.StartedAt, &e.EndedAt, &e.TraceCount); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// DeleteAllEvents removes every row in the events table; the FK cascade on
// traces.event_id wipes the traces table as a side effect.
func (s *Store) DeleteAllEvents(ctx context.Context) error {
	if s.db == nil {
		return errNilDB
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM events`)
	return err
}

// DeleteEventByID removes one event and (via FK cascade) its traces. Returns
// ErrEventNotFound when no row matches.
func (s *Store) DeleteEventByID(ctx context.Context, eventID string) error {
	if s.db == nil {
		return errNilDB
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return ErrEventNotFound
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE event_id = $1`, eventID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrEventNotFound
	}
	return nil
}

// ListTracesByEvent returns traces for an event in chronological order.
// limit caps the page size; offset is the row offset.
func (s *Store) ListTracesByEvent(ctx context.Context, eventID string, limit, offset int) ([]types.Trace, error) {
	if s.db == nil {
		return nil, errNilDB
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id::text, event_id, occurred_at, data
FROM traces
WHERE event_id = $1
ORDER BY occurred_at, id
LIMIT $2 OFFSET $3`, eventID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []types.Trace{}
	for rows.Next() {
		var t types.Trace
		var raw []byte
		if err := rows.Scan(&t.ID, &t.EventID, &t.OccurredAt, &raw); err != nil {
			return nil, err
		}
		t.Data = json.RawMessage(raw)
		out = append(out, t)
	}
	return out, rows.Err()
}
