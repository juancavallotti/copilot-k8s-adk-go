package recipes

import (
	"context"

	types "juancavallotti.com/recipe-types"
)

func (s *Service) CreateRecipe(ctx context.Context, recipe types.Recipe) (string, error) {
	if err := ValidateRecipeForCreate(recipe); err != nil {
		return "", err
	}
	return s.store.CreateRecipe(ctx, recipe)
}
