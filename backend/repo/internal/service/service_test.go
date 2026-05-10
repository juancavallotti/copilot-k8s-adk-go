package service

import (
	"context"
	"errors"
	"testing"

	types "juancavallotti.com/recipe-types"
)

type fakeStore struct {
	getRecipesCalls int
	getRecipeCalls  int
	createCalls     int
	updateCalls     int
	deleteCalls     int

	createErr error
}

func (f *fakeStore) GetRecipes(ctx context.Context) ([]types.Recipe, error) {
	f.getRecipesCalls++
	return nil, nil
}

func (f *fakeStore) GetRecipe(ctx context.Context, id string) (types.Recipe, error) {
	f.getRecipeCalls++
	return types.Recipe{ID: id, Name: "from-store"}, nil
}

func (f *fakeStore) CreateRecipe(ctx context.Context, recipe types.Recipe) (string, error) {
	f.createCalls++
	if f.createErr != nil {
		return "", f.createErr
	}
	return "new-recipe-id", nil
}

func (f *fakeStore) UpdateRecipe(ctx context.Context, recipe types.Recipe) error {
	f.updateCalls++
	return nil
}

func (f *fakeStore) DeleteRecipe(ctx context.Context, id string) error {
	f.deleteCalls++
	return nil
}

func TestService_GetRecipes_NoValidation(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	_, err := s.GetRecipes(context.Background())
	if err != nil {
		t.Fatalf("GetRecipes: %v", err)
	}
	if f.getRecipesCalls != 1 {
		t.Fatalf("store calls = %d, want 1", f.getRecipesCalls)
	}
}

func TestService_GetRecipe_ValidationShortCircuit(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	_, err := s.GetRecipe(context.Background(), "  ")
	if !errors.Is(err, ErrInvalidRecipeID) {
		t.Fatalf("err = %v, want ErrInvalidRecipeID", err)
	}
	if f.getRecipeCalls != 0 {
		t.Fatalf("store should not be called, calls=%d", f.getRecipeCalls)
	}
}

func TestService_GetRecipe_DelegatesToStore(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	got, err := s.GetRecipe(context.Background(), "abc")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "abc" || got.Name != "from-store" {
		t.Fatalf("got %+v", got)
	}
	if f.getRecipeCalls != 1 {
		t.Fatalf("store calls = %d", f.getRecipeCalls)
	}
}

func TestService_CreateRecipe_ValidationShortCircuit(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	_, err := s.CreateRecipe(context.Background(), types.Recipe{Name: ""})
	if !errors.Is(err, ErrInvalidRecipe) {
		t.Fatalf("err = %v", err)
	}
	if f.createCalls != 0 {
		t.Fatalf("store should not be called")
	}
}

func TestService_CreateRecipe_DelegatesToStore(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	r := types.Recipe{Name: "n", Ingredients: []string{"i"}, Instructions: []string{"s"}}
	if _, err := s.CreateRecipe(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if f.createCalls != 1 {
		t.Fatalf("createCalls = %d", f.createCalls)
	}
}

func TestService_UpdateRecipe_ValidationShortCircuit(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	err := s.UpdateRecipe(context.Background(), types.Recipe{ID: ""})
	if !errors.Is(err, ErrInvalidRecipeID) {
		t.Fatalf("err = %v", err)
	}
	if f.updateCalls != 0 {
		t.Fatalf("store should not be called")
	}
}

func TestService_DeleteRecipe_ValidationShortCircuit(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	if err := s.DeleteRecipe(context.Background(), ""); !errors.Is(err, ErrInvalidRecipeID) {
		t.Fatalf("err = %v", err)
	}
	if f.deleteCalls != 0 {
		t.Fatalf("store should not be called")
	}
}

func TestService_CreateRecipe_StoreErrorPropagates(t *testing.T) {
	t.Parallel()
	want := errors.New("db down")
	f := &fakeStore{createErr: want}
	s := &Service{store: f}
	r := types.Recipe{Name: "n", Ingredients: []string{"i"}, Instructions: []string{"s"}}
	_, err := s.CreateRecipe(context.Background(), r)
	if !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}
}
