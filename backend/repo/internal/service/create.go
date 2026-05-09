package service

import (
	"context"

	types "juancavallotti.com/recipe-types"
)

func (s *Service) CreateRecipe(ctx context.Context, recipe types.Recipe) error {
	return s.store.CreateRecipe(ctx, recipe)
}
