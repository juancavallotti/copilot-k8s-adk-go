package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	types "juancavallotti.com/recipe-types"
)

type fakeRepo struct {
	recipes []types.Recipe

	createRecipeCalls int
	createdRecipe     types.Recipe

	getRecipeCalls []string

	updateRecipeCalls int
	updatedRecipe     types.Recipe

	addPhotoCalls int
	addedPhoto    types.Photo

	deletePhotoCalls int
	deletedRecipeID  string
	deletedPhotoID   string

	deleteRecipeCalls int
	deletedID         string

	setFeaturedPhotoCalls int
	featuredRecipeID      string
	featuredPhotoID       string

	importedRecipes []types.Recipe
}

func (f *fakeRepo) GetRecipes(ctx context.Context) ([]types.Recipe, error) {
	return f.recipes, nil
}

func (f *fakeRepo) GetRecipe(ctx context.Context, id string) (types.Recipe, error) {
	f.getRecipeCalls = append(f.getRecipeCalls, id)
	if f.updatedRecipe.ID == id {
		return f.updatedRecipe, nil
	}
	if f.createdRecipe.ID == id {
		return f.createdRecipe, nil
	}
	for _, rec := range f.recipes {
		if rec.ID == id {
			return rec, nil
		}
	}
	return types.Recipe{ID: id, Name: "from-get"}, nil
}

func (f *fakeRepo) CreateRecipe(ctx context.Context, recipe types.Recipe) (string, error) {
	f.createRecipeCalls++
	f.createdRecipe = recipe
	f.createdRecipe.ID = "created-id"
	return f.createdRecipe.ID, nil
}

func (f *fakeRepo) UpdateRecipe(ctx context.Context, recipe types.Recipe) error {
	f.updateRecipeCalls++
	f.updatedRecipe = recipe
	return nil
}

func (f *fakeRepo) AddRecipePhoto(ctx context.Context, recipeID string, photo types.Photo) (string, error) {
	f.addPhotoCalls++
	f.addedPhoto = photo
	if f.updatedRecipe.ID == "" {
		f.updatedRecipe = types.Recipe{ID: recipeID, Name: "with-photo", Photos: []types.Photo{photo}}
	}
	return "photo-id", nil
}

func (f *fakeRepo) DeleteRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	f.deletePhotoCalls++
	f.deletedRecipeID = recipeID
	f.deletedPhotoID = photoID
	f.updatedRecipe = types.Recipe{ID: recipeID, Name: "without-photo", Photos: []types.Photo{}}
	return nil
}

func (f *fakeRepo) DeleteRecipe(ctx context.Context, id string) error {
	f.deleteRecipeCalls++
	f.deletedID = id
	return nil
}

func (f *fakeRepo) SetFeaturedRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	f.setFeaturedPhotoCalls++
	f.featuredRecipeID = recipeID
	f.featuredPhotoID = photoID
	f.updatedRecipe = types.Recipe{ID: recipeID, Name: "with-featured-photo", Photos: []types.Photo{
		{ID: photoID, ImageBase64: "aW1n", Featured: true},
	}}
	return nil
}

func (f *fakeRepo) ImportRecipe(ctx context.Context, recipe types.Recipe) error {
	f.importedRecipes = append(f.importedRecipes, recipe)
	return nil
}

func testRunner(stdin string, repo RecipeRepo, factoryCalls *int) (Runner, *bytes.Buffer, *bytes.Buffer) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	r := NewRunner(strings.NewReader(stdin), &stdout, &stderr, func() (RecipeRepo, error) {
		*factoryCalls++
		return repo, nil
	})
	return r, &stdout, &stderr
}

func TestRun_NoArgsWritesUsageAndDoesNotOpenRepo(t *testing.T) {
	var factoryCalls int
	r, _, stderr := testRunner("", &fakeRepo{}, &factoryCalls)

	err := r.Run(context.Background(), nil)
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if factoryCalls != 0 {
		t.Fatalf("repo factory calls = %d, want 0", factoryCalls)
	}
	if !strings.Contains(stderr.String(), "Commands:") {
		t.Fatalf("stderr = %q, want usage", stderr.String())
	}
}

func TestRun_SchemaPrintsValidJSONAndDoesNotOpenRepo(t *testing.T) {
	var factoryCalls int
	r, stdout, stderr := testRunner("", &fakeRepo{}, &factoryCalls)

	if err := r.Run(context.Background(), []string{"schema"}); err != nil {
		t.Fatalf("Run schema: %v", err)
	}
	if factoryCalls != 0 {
		t.Fatalf("repo factory calls = %d, want 0", factoryCalls)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var schema map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &schema); err != nil {
		t.Fatalf("schema output is not JSON: %v\n%s", err, stdout.String())
	}
	if got := schema["$schema"]; got != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("$schema = %v", got)
	}
	defs, ok := schema["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("$defs missing or wrong type: %#v", schema["$defs"])
	}
	if _, ok := defs["createRecipe"]; !ok {
		t.Fatalf("createRecipe definition missing from %#v", defs)
	}
	if _, ok := defs["recipePatch"]; !ok {
		t.Fatalf("recipePatch definition missing from %#v", defs)
	}
}

func TestRun_CreateReadsOneRecipeObjectAndPrintsCreatedRecipe(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, stdout, _ := testRunner(`{
		"name": "Pancakes",
		"description": "Breakfast",
		"ingredients": ["1 cup flour"],
		"instructions": ["Mix"],
		"category": "breakfast",
		"image": "pancakes.jpg"
	}`, repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"create", "-"}); err != nil {
		t.Fatalf("Run create: %v", err)
	}
	if factoryCalls != 1 {
		t.Fatalf("repo factory calls = %d, want 1", factoryCalls)
	}
	if repo.createRecipeCalls != 1 {
		t.Fatalf("create calls = %d, want 1", repo.createRecipeCalls)
	}
	if repo.createdRecipe.ID != "created-id" {
		t.Fatalf("created id = %q, want created-id", repo.createdRecipe.ID)
	}
	if repo.createdRecipe.Name != "Pancakes" {
		t.Fatalf("created name = %q", repo.createdRecipe.Name)
	}

	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if out.ID != "created-id" || out.Name != "Pancakes" {
		t.Fatalf("output recipe = %#v", out)
	}
}

func TestRun_CreateRejectsUnknownFields(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, _, _ := testRunner(`{"id":"client-id","name":"Pancakes"}`, repo, &factoryCalls)

	err := r.Run(context.Background(), []string{"create", "-"})
	if err == nil {
		t.Fatal("Run create returned nil, want unknown field error")
	}
	if !strings.Contains(err.Error(), `unknown field "id"`) {
		t.Fatalf("err = %v, want unknown field id", err)
	}
	if repo.createRecipeCalls != 0 {
		t.Fatalf("create calls = %d, want 0", repo.createRecipeCalls)
	}
}

func TestRun_PatchMergesProvidedFieldsAndPrintsUpdatedRecipe(t *testing.T) {
	repo := &fakeRepo{
		recipes: []types.Recipe{{
			ID:           "recipe-1",
			Name:         "Old",
			Description:  "Keep",
			Ingredients:  []string{"1 cup flour"},
			Instructions: []string{"Mix"},
			Category:     "breakfast",
			Image:        "old.jpg",
		}},
	}
	var factoryCalls int
	r, stdout, _ := testRunner(`{"name":"New","ingredients":["2 cups flour"]}`, repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"patch", " recipe-1 ", "-"}); err != nil {
		t.Fatalf("Run patch: %v", err)
	}
	if repo.updateRecipeCalls != 1 {
		t.Fatalf("update calls = %d, want 1", repo.updateRecipeCalls)
	}
	if repo.updatedRecipe.ID != "recipe-1" {
		t.Fatalf("updated id = %q", repo.updatedRecipe.ID)
	}
	if repo.updatedRecipe.Name != "New" {
		t.Fatalf("updated name = %q", repo.updatedRecipe.Name)
	}
	if repo.updatedRecipe.Description != "Keep" {
		t.Fatalf("description = %q, want Keep", repo.updatedRecipe.Description)
	}
	if got := repo.updatedRecipe.Ingredients; len(got) != 1 || got[0] != "2 cups flour" {
		t.Fatalf("ingredients = %#v", got)
	}

	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if out.Name != "New" || out.Description != "Keep" {
		t.Fatalf("output recipe = %#v", out)
	}
}

func TestRun_PatchRejectsEmptyPatch(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, _, _ := testRunner(`{}`, repo, &factoryCalls)

	err := r.Run(context.Background(), []string{"patch", "recipe-1", "-"})
	if err == nil || err.Error() != "no fields to update" {
		t.Fatalf("err = %v, want no fields to update", err)
	}
	if repo.updateRecipeCalls != 0 {
		t.Fatalf("update calls = %d, want 0", repo.updateRecipeCalls)
	}
}

func TestRun_AddPhotoReadsFileAndPrintsUpdatedRecipe(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	tmp := t.TempDir() + "/photo.bin"
	if err := os.WriteFile(tmp, []byte("img"), 0o600); err != nil {
		t.Fatal(err)
	}
	r, stdout, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"add-photo", " recipe-1 ", tmp, "--featured"}); err != nil {
		t.Fatalf("Run add-photo: %v", err)
	}
	if repo.addPhotoCalls != 1 {
		t.Fatalf("add photo calls = %d, want 1", repo.addPhotoCalls)
	}
	if repo.addedPhoto.ImageBase64 != "aW1n" || !repo.addedPhoto.Featured {
		t.Fatalf("added photo = %#v", repo.addedPhoto)
	}
	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if out.ID != "recipe-1" || len(out.Photos) != 1 {
		t.Fatalf("output recipe = %#v", out)
	}
}

func TestRun_AddPhotoReadsBase64FromStdin(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, stdout, _ := testRunner(" aW1n\n", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"add-photo", " recipe-1 ", "-", "--featured"}); err != nil {
		t.Fatalf("Run add-photo: %v", err)
	}
	if repo.addPhotoCalls != 1 {
		t.Fatalf("add photo calls = %d, want 1", repo.addPhotoCalls)
	}
	if repo.addedPhoto.ImageBase64 != "aW1n" || !repo.addedPhoto.Featured {
		t.Fatalf("added photo = %#v", repo.addedPhoto)
	}
	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if out.ID != "recipe-1" || len(out.Photos) != 1 {
		t.Fatalf("output recipe = %#v", out)
	}
}

func TestRun_AddPhotoRejectsInvalidBase64FromStdin(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, _, _ := testRunner("not base64!", repo, &factoryCalls)

	err := r.Run(context.Background(), []string{"add-photo", "recipe-1", "-"})
	if err == nil || !strings.Contains(err.Error(), "invalid base64 image data") {
		t.Fatalf("err = %v, want invalid base64 image data", err)
	}
	if repo.addPhotoCalls != 0 {
		t.Fatalf("add photo calls = %d, want 0", repo.addPhotoCalls)
	}
}

func TestRun_DeletePhotoRemovesPhotoAndPrintsUpdatedRecipe(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, stdout, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"delete-photo", " recipe-1 ", " photo-1 "}); err != nil {
		t.Fatalf("Run delete-photo: %v", err)
	}
	if repo.deletePhotoCalls != 1 {
		t.Fatalf("delete photo calls = %d, want 1", repo.deletePhotoCalls)
	}
	if repo.deletedRecipeID != "recipe-1" || repo.deletedPhotoID != "photo-1" {
		t.Fatalf("deleted recipe/photo = %q/%q", repo.deletedRecipeID, repo.deletedPhotoID)
	}
	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if out.ID != "recipe-1" || len(out.Photos) != 0 {
		t.Fatalf("output recipe = %#v", out)
	}
}

func TestRun_DeleteRemovesRecipe(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, stdout, stderr := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"delete", " recipe-1 "}); err != nil {
		t.Fatalf("Run delete: %v", err)
	}
	if repo.deleteRecipeCalls != 1 {
		t.Fatalf("delete recipe calls = %d, want 1", repo.deleteRecipeCalls)
	}
	if repo.deletedID != "recipe-1" {
		t.Fatalf("deleted id = %q, want recipe-1", repo.deletedID)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRun_SetFeaturedPhotoMarksPhotoAndPrintsUpdatedRecipe(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, stdout, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"set-featured-photo", " recipe-1 ", " photo-1 "}); err != nil {
		t.Fatalf("Run set-featured-photo: %v", err)
	}
	if repo.setFeaturedPhotoCalls != 1 {
		t.Fatalf("set featured photo calls = %d, want 1", repo.setFeaturedPhotoCalls)
	}
	if repo.featuredRecipeID != "recipe-1" || repo.featuredPhotoID != "photo-1" {
		t.Fatalf("featured recipe/photo = %q/%q", repo.featuredRecipeID, repo.featuredPhotoID)
	}
	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if out.ID != "recipe-1" || len(out.Photos) != 1 || !out.Photos[0].Featured {
		t.Fatalf("output recipe = %#v", out)
	}
}

func TestRun_ImportReadsJSONLines(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, _, _ := testRunner("{\"id\":\"1\",\"name\":\"One\"}\n\n{\"id\":\"2\",\"name\":\"Two\"}\n", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"import", "-"}); err != nil {
		t.Fatalf("Run import: %v", err)
	}
	if len(repo.importedRecipes) != 2 {
		t.Fatalf("imported recipes = %d, want 2", len(repo.importedRecipes))
	}
	if repo.importedRecipes[0].ID != "1" || repo.importedRecipes[1].ID != "2" {
		t.Fatalf("imported recipes = %#v", repo.importedRecipes)
	}
}

func TestRun_ListPrintsTable(t *testing.T) {
	repo := &fakeRepo{
		recipes: []types.Recipe{
			{ID: "1", Name: "One"},
			{ID: "2", Name: "Two"},
		},
	}
	var factoryCalls int
	r, stdout, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"list"}); err != nil {
		t.Fatalf("Run list: %v", err)
	}
	got := stdout.String()
	for _, want := range []string{"ID", "TITLE", "1", "One", "2", "Two"} {
		if !strings.Contains(got, want) {
			t.Fatalf("list output = %q, want %q", got, want)
		}
	}
}
