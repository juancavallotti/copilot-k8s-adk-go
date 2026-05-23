package recipes

import (
	"context"
	"errors"
	"testing"

	types "juancavallotti.com/recipe-types"

	recipeops "juancavallotti.com/recipes-repo/internal/dbops/recipes"
)

type fakeStore struct {
	getRecipesCalls   int
	getRecipeCalls    int
	createCalls       int
	createWithIDCalls int
	updateCalls       int
	addPhotoCalls     int
	deletePhotoCalls  int
	setFeaturedCalls  int
	deleteCalls       int

	createErr         error
	getRecipeNotFound bool
}

func (f *fakeStore) GetRecipes(ctx context.Context) ([]types.Recipe, error) {
	f.getRecipesCalls++
	return nil, nil
}

func (f *fakeStore) GetRecipe(ctx context.Context, id string) (types.Recipe, error) {
	f.getRecipeCalls++
	if f.getRecipeNotFound {
		return types.Recipe{}, recipeops.ErrRecipeNotFound
	}
	return types.Recipe{ID: id, Name: "from-store"}, nil
}

func (f *fakeStore) CreateRecipe(ctx context.Context, recipe types.Recipe) (string, error) {
	f.createCalls++
	if f.createErr != nil {
		return "", f.createErr
	}
	return "new-recipe-id", nil
}

func (f *fakeStore) CreateRecipeWithID(ctx context.Context, recipe types.Recipe) error {
	f.createWithIDCalls++
	return nil
}

func (f *fakeStore) UpdateRecipe(ctx context.Context, recipe types.Recipe) error {
	f.updateCalls++
	return nil
}

func (f *fakeStore) AddRecipePhoto(ctx context.Context, recipeID string, photo types.Photo) (string, error) {
	f.addPhotoCalls++
	return "photo-id", nil
}

func (f *fakeStore) DeleteRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	f.deletePhotoCalls++
	return nil
}

func (f *fakeStore) SetFeaturedRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	f.setFeaturedCalls++
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

func TestService_AddRecipePhoto_ValidatesBase64(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	_, err := s.AddRecipePhoto(context.Background(), testUUID, types.Photo{ImageBase64: "not base64"})
	if !errors.Is(err, ErrInvalidRecipe) {
		t.Fatalf("err = %v, want ErrInvalidRecipe", err)
	}
	if f.addPhotoCalls != 0 {
		t.Fatalf("store should not be called")
	}
}

func TestService_AddRecipePhoto_DelegatesToStore(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	id, err := s.AddRecipePhoto(context.Background(), testUUID, types.Photo{ImageBase64: "aW1n", Featured: true})
	if err != nil {
		t.Fatal(err)
	}
	if id != "photo-id" {
		t.Fatalf("id = %q, want photo-id", id)
	}
	if f.addPhotoCalls != 1 {
		t.Fatalf("addPhotoCalls = %d", f.addPhotoCalls)
	}
}

func TestService_DeleteRecipePhoto_ValidationShortCircuit(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	err := s.DeleteRecipePhoto(context.Background(), testUUID, " ")
	if !errors.Is(err, ErrInvalidRecipeID) {
		t.Fatalf("err = %v, want ErrInvalidRecipeID", err)
	}
	if f.deletePhotoCalls != 0 {
		t.Fatalf("store should not be called")
	}
}

func TestService_DeleteRecipePhoto_DelegatesToStore(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	if err := s.DeleteRecipePhoto(context.Background(), testUUID, testUUID); err != nil {
		t.Fatal(err)
	}
	if f.deletePhotoCalls != 1 {
		t.Fatalf("deletePhotoCalls = %d", f.deletePhotoCalls)
	}
}

func TestService_SetFeaturedRecipePhoto_ValidationShortCircuit(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	err := s.SetFeaturedRecipePhoto(context.Background(), testUUID, " ")
	if !errors.Is(err, ErrInvalidRecipeID) {
		t.Fatalf("err = %v, want ErrInvalidRecipeID", err)
	}
	if f.setFeaturedCalls != 0 {
		t.Fatalf("store should not be called")
	}
}

func TestService_SetFeaturedRecipePhoto_DelegatesToStore(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	if err := s.SetFeaturedRecipePhoto(context.Background(), testUUID, testUUID); err != nil {
		t.Fatal(err)
	}
	if f.setFeaturedCalls != 1 {
		t.Fatalf("setFeaturedCalls = %d", f.setFeaturedCalls)
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

const testUUID = "550e8400-e29b-41d4-a716-446655440000"

func TestService_ImportRecipe_EmptyID_Creates(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	r := types.Recipe{Name: "n", Ingredients: []string{"i"}, Instructions: []string{"s"}}
	if err := s.ImportRecipe(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if f.createCalls != 1 || f.getRecipeCalls != 0 || f.createWithIDCalls != 0 || f.updateCalls != 0 {
		t.Fatalf("calls create=%d get=%d withID=%d update=%d", f.createCalls, f.getRecipeCalls, f.createWithIDCalls, f.updateCalls)
	}
}

func TestService_ImportRecipe_NewUUID_InsertsWithID(t *testing.T) {
	t.Parallel()
	f := &fakeStore{getRecipeNotFound: true}
	s := &Service{store: f}
	r := types.Recipe{
		ID: testUUID, Name: "n", Ingredients: []string{"i"}, Instructions: []string{"s"},
	}
	if err := s.ImportRecipe(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if f.createWithIDCalls != 1 || f.updateCalls != 0 {
		t.Fatalf("withID=%d update=%d", f.createWithIDCalls, f.updateCalls)
	}
	if f.getRecipeCalls != 1 {
		t.Fatalf("getRecipeCalls = %d", f.getRecipeCalls)
	}
}

func TestService_ImportRecipe_ExistingUUID_Updates(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	r := types.Recipe{
		ID: testUUID, Name: "n", Ingredients: []string{"i"}, Instructions: []string{"s"},
	}
	if err := s.ImportRecipe(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if f.updateCalls != 1 || f.createWithIDCalls != 0 {
		t.Fatalf("withID=%d update=%d", f.createWithIDCalls, f.updateCalls)
	}
}
