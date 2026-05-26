package recipes

import (
	"context"

	types "juancavallotti.com/recipe-types"
	recipeops "juancavallotti.com/recipes-repo/internal/dbops/recipes"
)

type Service struct {
	store *recipeops.Store
}

// NewService wires a recipe store into the recipe service layer.
func NewService(store *recipeops.Store) *Service {
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

// SearchRecipes runs a semantic search and returns ranked recipes.
func (s *Service) SearchRecipes(ctx context.Context, query string, limit int) ([]types.RecipeMatch, error) {
	return s.store.SearchRecipes(ctx, query, limit)
}

// SearchRecipeChunks runs the same semantic search but returns the slim
// per-recipe form (id, name, best chunk, score). Preferred by agent
// callers that would otherwise pull photo base64 through their context.
func (s *Service) SearchRecipeChunks(ctx context.Context, query string, limit int) ([]types.RecipeHit, error) {
	return s.store.SearchRecipeChunks(ctx, query, limit)
}

// Wait blocks until in-flight async embedding work in the store has
// completed. Forwarded so the Repo can drain on shutdown.
func (s *Service) Wait() {
	s.store.Wait()
}
