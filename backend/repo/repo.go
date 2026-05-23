package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"

	types "juancavallotti.com/recipe-types"
	"juancavallotti.com/recipes-repo/internal/dbops"
	"juancavallotti.com/recipes-repo/internal/service"
)

type Repo struct {
	service *service.Service
	pool    *sql.DB
}

func (r *Repo) Ping(ctx context.Context) error {
	return r.pool.PingContext(ctx)
}

func (r *Repo) GetRecipes(ctx context.Context) ([]types.Recipe, error) {
	return r.service.GetRecipes(ctx)
}

func (r *Repo) GetRecipe(ctx context.Context, id string) (types.Recipe, error) {
	return r.service.GetRecipe(ctx, id)
}

func (r *Repo) CreateRecipe(ctx context.Context, recipe types.Recipe) (string, error) {
	return r.service.CreateRecipe(ctx, recipe)
}

func (r *Repo) UpdateRecipe(ctx context.Context, recipe types.Recipe) error {
	return r.service.UpdateRecipe(ctx, recipe)
}

func (r *Repo) AddRecipePhoto(ctx context.Context, recipeID string, photo types.Photo) (string, error) {
	return r.service.AddRecipePhoto(ctx, recipeID, photo)
}

func (r *Repo) DeleteRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	return r.service.DeleteRecipePhoto(ctx, recipeID, photoID)
}

func (r *Repo) SetFeaturedRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	return r.service.SetFeaturedRecipePhoto(ctx, recipeID, photoID)
}

func (r *Repo) DeleteRecipe(ctx context.Context, id string) error {
	return r.service.DeleteRecipe(ctx, id)
}

func (r *Repo) ImportRecipe(ctx context.Context, recipe types.Recipe) error {
	return r.service.ImportRecipe(ctx, recipe)
}

func (r *Repo) LogTrace(ctx context.Context, eventID string, occurredAt time.Time, data json.RawMessage) error {
	return r.service.LogTrace(ctx, eventID, occurredAt, data)
}

func (r *Repo) ListEvents(ctx context.Context, limit, offset int) ([]types.Event, error) {
	return r.service.ListEvents(ctx, limit, offset)
}

func (r *Repo) ListTracesByEvent(ctx context.Context, eventID string, limit, offset int) ([]types.Trace, error) {
	return r.service.ListTracesByEvent(ctx, eventID, limit, offset)
}

func (r *Repo) ListSkills(ctx context.Context) ([]types.Skill, error) {
	return r.service.ListSkills(ctx)
}

func (r *Repo) GetSkill(ctx context.Context, id string) (types.Skill, error) {
	return r.service.GetSkill(ctx, id)
}

func (r *Repo) GetSkillByName(ctx context.Context, name string) (types.Skill, error) {
	return r.service.GetSkillByName(ctx, name)
}

func (r *Repo) CreateSkill(ctx context.Context, name, description, content string) (string, error) {
	return r.service.CreateSkill(ctx, name, description, content)
}

func (r *Repo) UpdateSkill(ctx context.Context, id, description, content string) error {
	return r.service.UpdateSkill(ctx, id, description, content)
}

func (r *Repo) DeleteSkill(ctx context.Context, id string) error {
	return r.service.DeleteSkill(ctx, id)
}

func NewRepo() (*Repo, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	pool, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName))
	if err != nil {
		return nil, err
	}

	store := dbops.NewStore(pool)
	return &Repo{service: service.NewService(store), pool: pool}, nil
}
