package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"

	types "juancavallotti.com/recipe-types"
	recipeops "juancavallotti.com/recipes-repo/internal/dbops/recipes"
	skillops "juancavallotti.com/recipes-repo/internal/dbops/skills"
	traceops "juancavallotti.com/recipes-repo/internal/dbops/traces"
	"juancavallotti.com/recipes-repo/internal/embeddings"
	recipesvc "juancavallotti.com/recipes-repo/internal/service/recipes"
	skillsvc "juancavallotti.com/recipes-repo/internal/service/skills"
	tracesvc "juancavallotti.com/recipes-repo/internal/service/traces"
)

// Re-exports of internal types so external modules (cli, api) can use
// them without reaching into internal packages.
type (
	ReindexOptions       = recipeops.ReindexOptions
	IndexRecipeReport    = recipeops.IndexRecipeReport
	ReindexEventsOptions = traceops.ReindexEventsOptions
	IndexEventReport     = traceops.IndexEventReport
)

// ErrSearchDisabled is returned by Search* when no embedding API key
// is configured. Re-exported so HTTP handlers and the CLI can branch
// on it without importing the internal embeddings package.
var ErrSearchDisabled = embeddings.ErrDisabled

type Repo struct {
	recipes    *recipesvc.Service
	traces     *tracesvc.Service
	skills     *skillsvc.Service
	embeddings embeddings.Client
	pool       *sql.DB
}

// Embed produces a vector embedding for the given text. Returns
// embeddings.ErrDisabled when no API key is configured.
func (r *Repo) Embed(ctx context.Context, text string) ([]float32, error) {
	return r.embeddings.Embed(ctx, text)
}

func (r *Repo) Ping(ctx context.Context) error {
	return r.pool.PingContext(ctx)
}

// Close drains in-flight async embedding goroutines (recipes and
// traces) and then closes the underlying *sql.DB pool. Callers should
// defer this — short-lived processes (CLI invocations) would otherwise
// orphan the goroutines fired by write hooks. Prefer to defer it
// exactly once per Repo; lib/pq errors on a second pool.Close.
func (r *Repo) Close() error {
	r.recipes.Wait()
	r.traces.Wait()
	return r.pool.Close()
}

func (r *Repo) GetRecipes(ctx context.Context) ([]types.Recipe, error) {
	return r.recipes.GetRecipes(ctx)
}

func (r *Repo) GetRecipe(ctx context.Context, id string) (types.Recipe, error) {
	return r.recipes.GetRecipe(ctx, id)
}

func (r *Repo) CreateRecipe(ctx context.Context, recipe types.Recipe) (string, error) {
	return r.recipes.CreateRecipe(ctx, recipe)
}

func (r *Repo) UpdateRecipe(ctx context.Context, recipe types.Recipe) error {
	return r.recipes.UpdateRecipe(ctx, recipe)
}

func (r *Repo) AddRecipePhoto(ctx context.Context, recipeID string, photo types.Photo) (string, error) {
	return r.recipes.AddRecipePhoto(ctx, recipeID, photo)
}

func (r *Repo) DeleteRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	return r.recipes.DeleteRecipePhoto(ctx, recipeID, photoID)
}

func (r *Repo) SetFeaturedRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	return r.recipes.SetFeaturedRecipePhoto(ctx, recipeID, photoID)
}

func (r *Repo) DeleteRecipe(ctx context.Context, id string) error {
	return r.recipes.DeleteRecipe(ctx, id)
}

func (r *Repo) ImportRecipe(ctx context.Context, recipe types.Recipe) error {
	return r.recipes.ImportRecipe(ctx, recipe)
}

// IndexRecipe rebuilds the embedding rows for one recipe.
func (r *Repo) IndexRecipe(ctx context.Context, id string) error {
	return r.recipes.IndexRecipe(ctx, id)
}

// ReindexRecipes streams a bulk reindex pass. Use this from the CLI
// (and the agent) to backfill or rebuild the embedding table.
func (r *Repo) ReindexRecipes(ctx context.Context, opts ReindexOptions) error {
	return r.recipes.ReindexRecipes(ctx, opts)
}

// IndexEvent rebuilds the embedding row for one event.
func (r *Repo) IndexEvent(ctx context.Context, eventID string, force bool) error {
	return r.traces.IndexEvent(ctx, eventID, force)
}

// ReindexEvents streams a bulk reindex of events whose user_prompt
// is populated.
func (r *Repo) ReindexEvents(ctx context.Context, opts ReindexEventsOptions) error {
	return r.traces.ReindexEvents(ctx, opts)
}

// SearchRecipes runs a semantic-similarity search over recipe
// embeddings and returns matches ranked by best chunk score.
func (r *Repo) SearchRecipes(ctx context.Context, query string, limit int) ([]types.RecipeMatch, error) {
	return r.recipes.SearchRecipes(ctx, query, limit)
}

// SearchRecipeChunks is the slim variant of SearchRecipes: it returns
// id/name/best-chunk/score per hit rather than the full Recipe. Used by
// the CLI so an agent doesn't pull photo base64 through its context.
func (r *Repo) SearchRecipeChunks(ctx context.Context, query string, limit int) ([]types.RecipeHit, error) {
	return r.recipes.SearchRecipeChunks(ctx, query, limit)
}

// SearchEvents runs a semantic-similarity search over event
// embeddings (one per event, keyed off user_prompt).
func (r *Repo) SearchEvents(ctx context.Context, query string, limit int) ([]types.EventMatch, error) {
	return r.traces.SearchEvents(ctx, query, limit)
}

func (r *Repo) LogTrace(ctx context.Context, eventID string, occurredAt time.Time, data json.RawMessage) error {
	return r.traces.LogTrace(ctx, eventID, occurredAt, data)
}

func (r *Repo) ListEvents(ctx context.Context, limit, offset int) ([]types.Event, error) {
	return r.traces.ListEvents(ctx, limit, offset)
}

func (r *Repo) ListTracesByEvent(ctx context.Context, eventID string, limit, offset int) ([]types.Trace, error) {
	return r.traces.ListTracesByEvent(ctx, eventID, limit, offset)
}

func (r *Repo) DeleteAllEvents(ctx context.Context) error {
	return r.traces.DeleteAllEvents(ctx)
}

func (r *Repo) DeleteEvent(ctx context.Context, eventID string) error {
	return r.traces.DeleteEvent(ctx, eventID)
}

func (r *Repo) ListSkills(ctx context.Context) ([]types.Skill, error) {
	return r.skills.ListSkills(ctx)
}

func (r *Repo) GetSkill(ctx context.Context, id string) (types.Skill, error) {
	return r.skills.GetSkill(ctx, id)
}

func (r *Repo) GetSkillByName(ctx context.Context, name string) (types.Skill, error) {
	return r.skills.GetSkillByName(ctx, name)
}

func (r *Repo) CreateSkill(ctx context.Context, name, description, content string) (string, error) {
	return r.skills.CreateSkill(ctx, name, description, content)
}

func (r *Repo) UpdateSkill(ctx context.Context, id, description, content string) error {
	return r.skills.UpdateSkill(ctx, id, description, content)
}

func (r *Repo) DeleteSkill(ctx context.Context, id string) error {
	return r.skills.DeleteSkill(ctx, id)
}

func NewRepo() (*Repo, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	slog.Info("repo.opening", "host", dbHost, "port", dbPort, "database", dbName, "user", dbUser)
	pool, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName))
	if err != nil {
		slog.Error("repo.open_failed", "host", dbHost, "port", dbPort, "database", dbName, "user", dbUser, "err", err)
		return nil, err
	}

	embedClient, embedProvider := embeddings.NewFromEnv()
	// The dbops stores treat a nil embedder as "disabled" and skip
	// async indexing entirely, so we hand them nil rather than the
	// Noop fallback when no API key is configured.
	var storeEmbed embeddings.Client
	if embedProvider == embeddings.ProviderNoop {
		slog.Info("repo.embeddings_disabled", "reason", "no API key configured (set GEMINI_API_KEY or OPENAI_API_KEY)")
	} else {
		storeEmbed = embedClient
		slog.Info("repo.embeddings_enabled", "provider", string(embedProvider), "dims", embeddings.Dimensions)
	}

	slog.Info("repo.initialized", "database", dbName)
	return &Repo{
		recipes:    recipesvc.NewService(recipeops.NewStore(pool, recipeops.WithEmbedClient(storeEmbed))),
		traces:     tracesvc.NewService(traceops.NewStore(pool, traceops.WithEmbedClient(storeEmbed))),
		skills:     skillsvc.NewService(skillops.NewStore(pool)),
		embeddings: embedClient,
		pool:       pool,
	}, nil
}
