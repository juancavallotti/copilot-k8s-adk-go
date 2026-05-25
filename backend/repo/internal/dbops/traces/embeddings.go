package traces

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	types "juancavallotti.com/recipe-types"
	"juancavallotti.com/recipes-repo/internal/embeddings"
)

// embedTimeout caps how long a single async embedding write can run
// before its goroutine gives up.
const embedTimeout = 60 * time.Second

// IndexEventReport is one row in the reindex output stream — same
// shape as recipes' IndexRecipeReport so the CLI can render either.
type IndexEventReport struct {
	ID     string `json:"id"`
	Status string `json:"status"`          // "ok" | "error" | "skipped"
	Error  string `json:"error,omitempty"` // populated when Status == "error"
}

// ReindexEventsOptions controls a bulk event-reindex pass.
type ReindexEventsOptions struct {
	// Force re-embeds events that already have an embedding row.
	// Default behaviour is to skip them.
	Force bool
	// Limit caps the number of events processed. 0 means no limit.
	Limit int
	// OnReport is called for every processed event. The reindex
	// command streams these to stdout.
	OnReport func(IndexEventReport)
}

// IndexEvent embeds the event's user_prompt and writes its row in
// event_embeddings. When force is false, returns nil without calling
// the embedder if a row already exists for this event — that's the
// path used by the InsertTrace write hook, which fires on every
// trace insert and shouldn't repeatedly re-embed the same prompt.
func (s *Store) IndexEvent(ctx context.Context, eventID string, force bool) error {
	if s.db == nil {
		return errNilDB
	}
	if _, disabled := s.embed.(embeddings.Noop); disabled {
		return embeddings.ErrDisabled
	}
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return errors.New("traces.IndexEvent: empty eventID")
	}

	var prompt sql.NullString
	if err := s.db.QueryRowContext(ctx, `SELECT user_prompt FROM events WHERE event_id = $1`, eventID).Scan(&prompt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEventNotFound
		}
		return err
	}
	text := strings.TrimSpace(prompt.String)
	if text == "" {
		return nil // nothing to embed; not an error
	}

	if !force {
		var exists bool
		if err := s.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM event_embeddings WHERE event_id = $1)`, eventID,
		).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return nil
		}
	}

	vec, err := s.embed.Embed(ctx, text)
	if err != nil {
		return fmt.Errorf("embed event: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO event_embeddings (event_id, source_text, embedding)
VALUES ($1, $2, $3::vector)
ON CONFLICT (event_id) DO UPDATE SET
    source_text = EXCLUDED.source_text,
    embedding   = EXCLUDED.embedding,
    updated_at  = now()`,
		eventID, text, embeddings.FormatVector(vec),
	); err != nil {
		return err
	}
	return nil
}

// ReindexEvents iterates events whose user_prompt is non-empty and
// indexes each one. When opts.Force is false, only events that don't
// already have an event_embeddings row are processed.
func (s *Store) ReindexEvents(ctx context.Context, opts ReindexEventsOptions) error {
	if s.db == nil {
		return errNilDB
	}
	q := `SELECT event_id FROM events WHERE user_prompt IS NOT NULL AND user_prompt <> ''`
	if !opts.Force {
		q += ` AND NOT EXISTS (SELECT 1 FROM event_embeddings ee WHERE ee.event_id = events.event_id)`
	}
	q += ` ORDER BY started_at`
	if opts.Limit > 0 {
		q += fmt.Sprintf(` LIMIT %d`, opts.Limit)
	}

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return err
	}
	ids := make([]string, 0, 64)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		rep := IndexEventReport{ID: id, Status: "ok"}
		if err := s.IndexEvent(ctx, id, opts.Force); err != nil {
			rep.Status = "error"
			rep.Error = err.Error()
		}
		if opts.OnReport != nil {
			opts.OnReport(rep)
		}
	}
	return nil
}

// SearchEvents runs a semantic-similarity search over the event
// embeddings. One row per event in event_embeddings, so no GROUP BY
// — the index ORDER BY uses the HNSW vector index directly.
func (s *Store) SearchEvents(ctx context.Context, query string, limit int) ([]types.EventMatch, error) {
	if s.db == nil {
		return nil, errNilDB
	}
	if _, disabled := s.embed.(embeddings.Noop); disabled {
		return nil, embeddings.ErrDisabled
	}
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("dbops/traces.SearchEvents: empty query")
	}
	if limit <= 0 {
		limit = 10
	}

	vec, err := s.embed.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT e.event_id, e.started_at, e.ended_at, e.trace_count, COALESCE(e.user_prompt, ''),
       1 - (ee.embedding <=> $1::vector) AS score
FROM event_embeddings ee
JOIN events e ON e.event_id = ee.event_id
ORDER BY ee.embedding <=> $1::vector
LIMIT $2`,
		embeddings.FormatVector(vec), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]types.EventMatch, 0, limit)
	for rows.Next() {
		var m types.EventMatch
		if err := rows.Scan(&m.EventID, &m.StartedAt, &m.EndedAt, &m.TraceCount, &m.UserPrompt, &m.Score); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// indexEventAsync fires an embedding for eventID in a background
// goroutine. Called by InsertTrace after commit when the trace's
// user_prompt is non-empty. No-op when the embedding client is a
// no-op. The goroutine is tracked via s.wg so Store.Wait drains it.
func (s *Store) indexEventAsync(ctx context.Context, eventID string) {
	if _, disabled := s.embed.(embeddings.Noop); disabled {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), embedTimeout)
		defer cancel()
		if err := s.IndexEvent(bg, eventID, false); err != nil && !errors.Is(err, embeddings.ErrDisabled) {
			slog.Error("event.embedding_failed", "event_id", eventID, "err", err)
		}
	}()
}
