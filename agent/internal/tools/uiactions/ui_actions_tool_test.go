package uiactions

import "testing"

func TestNormalizeRequiresRecipeIDForNavigateRecipe(t *testing.T) {
	_, err := Normalize(Args{
		Actions: []Action{{Type: "navigate_recipe"}},
	})
	if err == nil {
		t.Fatal("Normalize() error = nil, want error")
	}
}

func TestNormalizeAllowsRefresh(t *testing.T) {
	result, err := Normalize(Args{
		Actions: []Action{{Type: " refresh_current_screen ", RecipeID: "ignored"}},
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if len(result.Actions) != 1 {
		t.Fatalf("Normalize() actions length = %d, want 1", len(result.Actions))
	}
	if got := result.Actions[0]; got.Type != "refresh_current_screen" || got.RecipeID != "" {
		t.Fatalf("Normalize() action = %+v, want refresh without recipe ID", got)
	}
}
