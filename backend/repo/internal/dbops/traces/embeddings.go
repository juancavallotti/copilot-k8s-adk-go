package traces

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

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
