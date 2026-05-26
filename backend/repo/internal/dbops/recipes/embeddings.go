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
// Returns embeddings.ErrDisabled when no embedder is wired in so the
// caller can decide whether that's an error or expected.
func (s *Store) IndexRecipe(ctx context.Context, id string) error {
	if s.db == nil {
		return errNilDB
	}
	if s.embed == nil {
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
// goroutine. Called from write hooks. No-op when no embedder is wired
// in, so dev environments without an API key don't log every recipe
// write. The goroutine is tracked via s.wg so Store.Wait can drain
// in-flight work before the process exits.
func (s *Store) indexRecipeAsync(ctx context.Context, id string) {
	if s.embed == nil {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
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

// SearchRecipes runs a semantic-similarity search over the recipe
// embeddings, hydrating each match into a full Recipe. Because each
// recipe owns multiple chunks (summary / ingredients / directions),
// the SQL groups by recipe_id and takes the best chunk score per
// recipe. Empty query is rejected — the embedder errors on it and
// the caller likely has a bug.
func (s *Store) SearchRecipes(ctx context.Context, query string, limit int) ([]types.RecipeMatch, error) {
	if s.db == nil {
		return nil, errNilDB
	}
	if s.embed == nil {
		return nil, embeddings.ErrDisabled
	}
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("dbops/recipes.SearchRecipes: empty query")
	}
	if limit <= 0 {
		limit = 10
	}

	vec, err := s.embed.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT recipe_id::text, MAX(1 - (embedding <=> $1::vector)) AS score
FROM recipe_embeddings
GROUP BY recipe_id
ORDER BY score DESC
LIMIT $2`,
		embeddings.FormatVector(vec), limit,
	)
	if err != nil {
		return nil, err
	}
	type hit struct {
		id    string
		score float64
	}
	hits := make([]hit, 0, limit)
	for rows.Next() {
		var h hit
		if err := rows.Scan(&h.id, &h.score); err != nil {
			rows.Close()
			return nil, err
		}
		hits = append(hits, h)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	matches := make([]types.RecipeMatch, 0, len(hits))
	for _, h := range hits {
		rec, err := s.GetRecipe(ctx, h.id)
		if err != nil {
			// Recipe may have been deleted between embedding and search;
			// silently drop those rather than failing the whole call.
			if errors.Is(err, ErrRecipeNotFound) {
				continue
			}
			return nil, err
		}
		matches = append(matches, types.RecipeMatch{Recipe: rec, Score: h.score})
	}
	return matches, nil
}

// SearchRecipeChunks runs the same semantic search as SearchRecipes but
// returns a slim per-recipe row: id, name, the best-matching chunk's
// source text, and the chunk's score. Used by agent-facing callers
// (the CLI) where pulling full Recipe objects — including photo
// base64 — would blow up context size.
func (s *Store) SearchRecipeChunks(ctx context.Context, query string, limit int) ([]types.RecipeHit, error) {
	if s.db == nil {
		return nil, errNilDB
	}
	if s.embed == nil {
		return nil, embeddings.ErrDisabled
	}
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("dbops/recipes.SearchRecipeChunks: empty query")
	}
	if limit <= 0 {
		limit = 10
	}

	vec, err := s.embed.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT id, name, chunk, score
FROM (
  SELECT DISTINCT ON (re.recipe_id)
    re.recipe_id::text AS id,
    r.name            AS name,
    re.source_text    AS chunk,
    1 - (re.embedding <=> $1::vector) AS score
  FROM recipe_embeddings re
  JOIN recipes r ON r.id = re.recipe_id
  ORDER BY re.recipe_id, re.embedding <=> $1::vector ASC
) best
ORDER BY score DESC
LIMIT $2`,
		embeddings.FormatVector(vec), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hits := make([]types.RecipeHit, 0, limit)
	for rows.Next() {
		var h types.RecipeHit
		if err := rows.Scan(&h.ID, &h.Name, &h.Chunk, &h.Score); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hits, nil
}
