package recipes

import (
	"context"

	types "juancavallotti.com/recipe-types"
)

func (s *Service) GetRecipes(ctx context.Context) ([]types.Recipe, error) {
	return s.store.GetRecipes(ctx)
}

func (s *Service) GetRecipe(ctx context.Context, id string) (types.Recipe, error) {
	if err := ValidateRecipeID(id); err != nil {
		return types.Recipe{}, err
	}
	return s.store.GetRecipe(ctx, id)
}
