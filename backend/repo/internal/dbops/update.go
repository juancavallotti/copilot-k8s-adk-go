package dbops

import (
	"context"

	types "juancavallotti.com/recipe-types"
)

func (s *Store) UpdateRecipe(ctx context.Context, recipe types.Recipe) error {
	if s.db == nil {
		return errNilDB
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	_ = recipe
	return nil
}
