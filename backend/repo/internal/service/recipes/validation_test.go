package recipes

import (
	"errors"
	"testing"

	types "juancavallotti.com/recipe-types"
)

func TestValidateRecipeID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{"ok", "abc-123", nil},
		{"trimmed ok", "  x  ", nil},
		{"empty", "", ErrInvalidRecipeID},
		{"whitespace only", "   \t\n", ErrInvalidRecipeID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateRecipeID(tt.id)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want errors.Is(..., %v)", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRecipeForCreate(t *testing.T) {
	t.Parallel()
	valid := types.Recipe{
		Name:         "Soup",
		Ingredients:  []string{"water"},
		Instructions: []string{"boil"},
	}
	tests := []struct {
		name    string
		recipe  types.Recipe
		wantErr error
	}{
		{"ok", valid, nil},
		{"ok with optional fields", types.Recipe{
			Name: "x", Description: "d", Category: "c", Image: "i",
			Ingredients: []string{"a"}, Instructions: []string{"b"},
		}, nil},
		{"id set", types.Recipe{ID: "1", Name: "x", Ingredients: []string{"a"}, Instructions: []string{"b"}}, ErrInvalidRecipe},
		{"empty name", types.Recipe{Name: "  ", Ingredients: []string{"a"}, Instructions: []string{"b"}}, ErrInvalidRecipe},
		{"no ingredients", types.Recipe{Name: "x", Ingredients: nil, Instructions: []string{"b"}}, ErrInvalidRecipe},
		{"only blank ingredients", types.Recipe{Name: "x", Ingredients: []string{" ", ""}, Instructions: []string{"b"}}, ErrInvalidRecipe},
		{"no instructions", types.Recipe{Name: "x", Ingredients: []string{"a"}, Instructions: nil}, ErrInvalidRecipe},
		{"only blank instructions", types.Recipe{Name: "x", Ingredients: []string{"a"}, Instructions: []string{"", "\t"}}, ErrInvalidRecipe},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateRecipeForCreate(tt.recipe)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want errors.Is(..., %v)", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRecipeForUpdate(t *testing.T) {
	t.Parallel()
	valid := types.Recipe{
		ID:           "r1",
		Name:         "Soup",
		Ingredients:  []string{"water"},
		Instructions: []string{"boil"},
	}
	tests := []struct {
		name    string
		recipe  types.Recipe
		wantErr error
	}{
		{"ok", valid, nil},
		{"missing id", types.Recipe{Name: "x", Ingredients: []string{"a"}, Instructions: []string{"b"}}, ErrInvalidRecipeID},
		{"blank id", types.Recipe{ID: "  ", Name: "x", Ingredients: []string{"a"}, Instructions: []string{"b"}}, ErrInvalidRecipeID},
		{"bad content", types.Recipe{ID: "1", Name: "", Ingredients: []string{"a"}, Instructions: []string{"b"}}, ErrInvalidRecipe},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateRecipeForUpdate(tt.recipe)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want errors.Is(..., %v)", err, tt.wantErr)
			}
		})
	}
}
