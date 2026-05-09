package service

import (
	"context"
)

func (s *Service) DeleteRecipe(ctx context.Context, id string) error {
	return s.store.DeleteRecipe(ctx, id)
}
