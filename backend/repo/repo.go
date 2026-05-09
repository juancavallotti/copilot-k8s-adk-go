package repo

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	types "juancavallotti.com/recipe-types"
	"juancavallotti.com/recipes-repo/internal/dbops"
	"juancavallotti.com/recipes-repo/internal/service"
)

type Repo struct {
	service *service.Service
}

func (r *Repo) GetRecipes(ctx context.Context) ([]types.Recipe, error) {
	return r.service.GetRecipes(ctx)
}

func (r *Repo) GetRecipe(ctx context.Context, id string) (types.Recipe, error) {
	return r.service.GetRecipe(ctx, id)
}

func (r *Repo) CreateRecipe(ctx context.Context, recipe types.Recipe) error {
	return r.service.CreateRecipe(ctx, recipe)
}

func (r *Repo) UpdateRecipe(ctx context.Context, recipe types.Recipe) error {
	return r.service.UpdateRecipe(ctx, recipe)
}

func (r *Repo) DeleteRecipe(ctx context.Context, id string) error {
	return r.service.DeleteRecipe(ctx, id)
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

	store := dbops.NewStore(pool, "recipes")
	return &Repo{service: service.NewService(store)}, nil
}
