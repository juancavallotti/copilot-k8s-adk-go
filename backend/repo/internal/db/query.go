package db

import (
	"context"

	types "juancavallotti.com/recipe-types"
)

func (s *Store) GetRecipes(ctx context.Context) ([]types.Recipe, error) {
	if s.db == nil {
		return nil, errNilDB
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Store) GetRecipe(ctx context.Context, id string) (types.Recipe, error) {
	if s.db == nil {
		return types.Recipe{}, errNilDB
	}
	if err := ctx.Err(); err != nil {
		return types.Recipe{}, err
	}
	_ = id
	return types.Recipe{}, nil
}
