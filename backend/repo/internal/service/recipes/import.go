package recipes

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	types "juancavallotti.com/recipe-types"

	recipeops "juancavallotti.com/recipes-repo/internal/dbops/recipes"
)

// ImportRecipe creates a new recipe when recipe.ID is empty, or upserts when ID is a UUID
// (updates if the row exists, otherwise inserts with that id). Backup JSONL from export-all
// can be replayed with this method.
func (s *Service) ImportRecipe(ctx context.Context, recipe types.Recipe) error {
	id := strings.TrimSpace(recipe.ID)
	if id == "" {
		_, err := s.CreateRecipe(ctx, recipe)
		return err
	}
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRecipeID, err)
	}
	if err := validateRecipeContent(recipe); err != nil {
		return err
	}
	_, err := s.store.GetRecipe(ctx, id)
	if err == nil {
		return s.store.UpdateRecipe(ctx, recipe)
	}
	if !errors.Is(err, recipeops.ErrRecipeNotFound) {
		return err
	}
	return s.store.CreateRecipeWithID(ctx, recipe)
}
