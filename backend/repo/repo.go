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
	recipesvc "juancavallotti.com/recipes-repo/internal/service/recipes"
	skillsvc "juancavallotti.com/recipes-repo/internal/service/skills"
	tracesvc "juancavallotti.com/recipes-repo/internal/service/traces"
)

type Repo struct {
	recipes *recipesvc.Service
	traces  *tracesvc.Service
	skills  *skillsvc.Service
	pool    *sql.DB
}

func (r *Repo) Ping(ctx context.Context) error {
	return r.pool.PingContext(ctx)
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

	slog.Info("repo.initialized", "database", dbName)
	return &Repo{
		recipes: recipesvc.NewService(recipeops.NewStore(pool)),
		traces:  tracesvc.NewService(traceops.NewStore(pool)),
		skills:  skillsvc.NewService(skillops.NewStore(pool)),
		pool:    pool,
	}, nil
}
