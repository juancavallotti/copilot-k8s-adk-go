package service

import (
	"context"

	types "juancavallotti.com/recipe-types"
)

func (s *Service) UpdateRecipe(ctx context.Context, recipe types.Recipe) error {
	return s.store.UpdateRecipe(ctx, recipe)
}
