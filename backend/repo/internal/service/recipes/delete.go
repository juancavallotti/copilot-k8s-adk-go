package recipes

import "context"

func (s *Service) DeleteRecipe(ctx context.Context, id string) error {
	if err := ValidateRecipeID(id); err != nil {
		return err
	}
	return s.store.DeleteRecipe(ctx, id)
}
