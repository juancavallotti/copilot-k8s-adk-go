package service

import (
	"context"

	types "juancavallotti.com/recipe-types"

	"juancavallotti.com/recipes-repo/internal/dbops"
)

// recipeStore is the persistence surface used by Service (implemented by *dbops.Store).
type recipeStore interface {
	GetRecipes(ctx context.Context) ([]types.Recipe, error)
	GetRecipe(ctx context.Context, id string) (types.Recipe, error)
	CreateRecipe(ctx context.Context, recipe types.Recipe) (string, error)
	CreateRecipeWithID(ctx context.Context, recipe types.Recipe) error
	UpdateRecipe(ctx context.Context, recipe types.Recipe) error
	AddRecipePhoto(ctx context.Context, recipeID string, photo types.Photo) (string, error)
	DeleteRecipePhoto(ctx context.Context, recipeID string, photoID string) error
	SetFeaturedRecipePhoto(ctx context.Context, recipeID string, photoID string) error
	DeleteRecipe(ctx context.Context, id string) error
}

type Service struct {
	store recipeStore
}

// NewService wires a db-backed store into the service layer.
func NewService(store *dbops.Store) *Service {
	return &Service{store: store}
}
