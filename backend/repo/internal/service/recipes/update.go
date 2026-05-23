package recipes

import (
	"context"

	types "juancavallotti.com/recipe-types"
)

func (s *Service) UpdateRecipe(ctx context.Context, recipe types.Recipe) error {
	if err := ValidateRecipeForUpdate(recipe); err != nil {
		return err
	}
	return s.store.UpdateRecipe(ctx, recipe)
}
