package dbops

import "context"

func (s *Store) DeleteRecipe(ctx context.Context, id string) error {
	if s.db == nil {
		return errNilDB
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	_ = id
	return nil
}
