package recipes

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	types "juancavallotti.com/recipe-types"
	"juancavallotti.com/recipes-repo/internal/embeddings"
)

// embedTimeout caps how long a single embedding write can run before
// the goroutine gives up — keeps a stalled Gemini/OpenAI call from
// pinning resources forever.
const embedTimeout = 60 * time.Second

// IndexRecipeReport is one row in the reindex output stream.
type IndexRecipeReport struct {
	ID     string `json:"id"`
	Status string `json:"status"`          // "ok" | "error" | "skipped"
	Error  string `json:"error,omitempty"` // populated when Status == "error"
}

// ReindexOptions controls a bulk reindex pass.
type ReindexOptions struct {
	// Force re-embeds rows that already have embeddings. Default behaviour
	// is to skip recipes that already have at least one embedding row.
	Force bool
	// Limit caps the number of recipes processed. 0 means no limit.
	Limit int
	// OnReport is called for every processed recipe. The reindex command
	// streams these to stdout so the agent can parse progress.
	OnReport func(IndexRecipeReport)
}

// IndexRecipe embeds the recipe's text chunks (summary, ingredients,
// directions) and replaces its rows in recipe_embeddings. Idempotent.
// Returns embeddings.ErrDisabled when the client is a no-op so the
// caller can decide whether that's an error or expected.
func (s *Store) IndexRecipe(ctx context.Context, id string) error {
	if s.db == nil {
		return errNilDB
	}
	if _, disabled := s.embed.(embeddings.Noop); disabled {
		return embeddings.ErrDisabled
	}
	rec, err := s.GetRecipe(ctx, id)
	if err != nil {
		return err
	}
	chunks := recipeChunks(rec)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM recipe_embeddings WHERE recipe_id = $1::uuid`, id); err != nil {
		return err
	}
	for _, text := range chunks {
		vec, err := s.embed.Embed(ctx, text)
		if err != nil {
			return fmt.Errorf("embed chunk: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO recipe_embeddings (recipe_id, source_text, embedding)
VALUES ($1::uuid, $2, $3::vector)`,
			id, text, embeddings.FormatVector(vec),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ReindexRecipes walks recipes and re-embeds them. With opts.Force=false
// it only processes recipes that have no existing embedding rows. Streams
// per-recipe outcomes through opts.OnReport.
func (s *Store) ReindexRecipes(ctx context.Context, opts ReindexOptions) error {
	if s.db == nil {
		return errNilDB
	}
	q := `SELECT id::text FROM recipes`
	if !opts.Force {
		q += ` WHERE NOT EXISTS (SELECT 1 FROM recipe_embeddings re WHERE re.recipe_id = recipes.id)`
	}
	q += ` ORDER BY created_at`
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
		rep := IndexRecipeReport{ID: id, Status: "ok"}
		if err := s.IndexRecipe(ctx, id); err != nil {
			rep.Status = "error"
			rep.Error = err.Error()
		}
		if opts.OnReport != nil {
			opts.OnReport(rep)
		}
	}
	return nil
}

// indexRecipeAsync fires an embedding rebuild for id in a background
// goroutine. Called from write hooks. No-op when the embedding client
// is a no-op so dev environments without an API key don't log every
// recipe write.
func (s *Store) indexRecipeAsync(ctx context.Context, id string) {
	if _, disabled := s.embed.(embeddings.Noop); disabled {
		return
	}
	go func() {
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), embedTimeout)
		defer cancel()
		if err := s.IndexRecipe(bg, id); err != nil && !errors.Is(err, embeddings.ErrDisabled) {
			slog.Error("recipe.embedding_failed", "id", id, "err", err)
		}
	}()
}

// recipeChunks builds the embedding source-texts for one recipe.
// Today there are up to three: summary (name + description), the
// ingredient list, and the directions. Empty chunks are dropped so
// the embedder isn't called with whitespace.
func recipeChunks(r types.Recipe) []string {
	chunks := make([]string, 0, 3)
	if summary := joinNonEmpty([]string{r.Name, r.Description}, "\n"); summary != "" {
		chunks = append(chunks, summary)
	}
	if ing := joinNonEmpty(r.Ingredients, "\n"); ing != "" {
		chunks = append(chunks, ing)
	}
	if dirs := joinNonEmpty(r.Instructions, "\n"); dirs != "" {
		chunks = append(chunks, dirs)
	}
	return chunks
}

func joinNonEmpty(lines []string, sep string) string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}
	return strings.Join(out, sep)
}
