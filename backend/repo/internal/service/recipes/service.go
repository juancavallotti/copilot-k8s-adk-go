package recipes

import (
	"context"

	types "juancavallotti.com/recipe-types"
	recipeops "juancavallotti.com/recipes-repo/internal/dbops/recipes"
)

type store interface {
	GetRecipes(ctx context.Context) ([]types.Recipe, error)
	GetRecipe(ctx context.Context, id string) (types.Recipe, error)
	CreateRecipe(ctx context.Context, recipe types.Recipe) (string, error)
	CreateRecipeWithID(ctx context.Context, recipe types.Recipe) error
	UpdateRecipe(ctx context.Context, recipe types.Recipe) error
	AddRecipePhoto(ctx context.Context, recipeID string, photo types.Photo) (string, error)
	DeleteRecipePhoto(ctx context.Context, recipeID string, photoID string) error
	SetFeaturedRecipePhoto(ctx context.Context, recipeID string, photoID string) error
	DeleteRecipe(ctx context.Context, id string) error
	IndexRecipe(ctx context.Context, id string) error
	ReindexRecipes(ctx context.Context, opts recipeops.ReindexOptions) error
}

type Service struct {
	store store
}

// NewService wires a recipe store into the recipe service layer.
func NewService(store store) *Service {
	return &Service{store: store}
}

// IndexRecipe rebuilds the embedding rows for a single recipe.
func (s *Service) IndexRecipe(ctx context.Context, id string) error {
	return s.store.IndexRecipe(ctx, id)
}

// ReindexRecipes streams a bulk reindex pass through the store. The
// service does no validation here — reindex is an ops-style operation,
// not a user-facing write.
func (s *Service) ReindexRecipes(ctx context.Context, opts recipeops.ReindexOptions) error {
	return s.store.ReindexRecipes(ctx, opts)
}
