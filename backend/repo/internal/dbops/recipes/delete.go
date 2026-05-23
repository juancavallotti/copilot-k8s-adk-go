package recipes

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

func (s *Store) DeleteRecipe(ctx context.Context, id string) error {
	if s.db == nil {
		return errNilDB
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	res, err := s.db.ExecContext(ctx, `DELETE FROM recipes WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrRecipeNotFound
	}
	return nil
}
