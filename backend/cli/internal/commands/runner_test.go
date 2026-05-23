package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

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

	logTraceCalls    int
	logTraceEntries  []traceEntry
	logTraceErr      error

	listEventsCalls  int
	listEventsLimit  int
	listEventsOffset int
	listEventsResult []types.Event
	listEventsErr    error

	listTracesCalls    int
	listTracesEventID  string
	listTracesLimit    int
	listTracesOffset   int
	listTracesResult   []types.Trace
	listTracesErr      error
}

type traceEntry struct {
	eventID    string
	occurredAt time.Time
	data       json.RawMessage
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

func (f *fakeRepo) LogTrace(ctx context.Context, eventID string, occurredAt time.Time, data json.RawMessage) error {
	f.logTraceCalls++
	f.logTraceEntries = append(f.logTraceEntries, traceEntry{eventID: eventID, occurredAt: occurredAt, data: data})
	return f.logTraceErr
}

func (f *fakeRepo) ListEvents(ctx context.Context, limit, offset int) ([]types.Event, error) {
	f.listEventsCalls++
	f.listEventsLimit = limit
	f.listEventsOffset = offset
	return f.listEventsResult, f.listEventsErr
}

func (f *fakeRepo) ListTracesByEvent(ctx context.Context, eventID string, limit, offset int) ([]types.Trace, error) {
	f.listTracesCalls++
	f.listTracesEventID = eventID
	f.listTracesLimit = limit
	f.listTracesOffset = offset
	return f.listTracesResult, f.listTracesErr
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

func TestRun_NoArgsPrintsHelpToStdoutAndDoesNotOpenRepo(t *testing.T) {
	var factoryCalls int
	r, stdout, stderr := testRunner("", &fakeRepo{}, &factoryCalls)

	err := r.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if factoryCalls != 0 {
		t.Fatalf("repo factory calls = %d, want 0", factoryCalls)
	}
	if !strings.Contains(stdout.String(), "Commands:") {
		t.Fatalf("stdout = %q, want help", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRun_HelpFlagsPrintHelpToStdout(t *testing.T) {
	for _, flag := range []string{"-h", "--help", "help"} {
		t.Run(flag, func(t *testing.T) {
			var factoryCalls int
			r, stdout, stderr := testRunner("", &fakeRepo{}, &factoryCalls)

			err := r.Run(context.Background(), []string{flag})
			if err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			if factoryCalls != 0 {
				t.Fatalf("repo factory calls = %d, want 0", factoryCalls)
			}
			if !strings.Contains(stdout.String(), "Commands:") {
				t.Fatalf("stdout = %q, want help", stdout.String())
			}
			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
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

func TestRun_CreateReadsOneRecipeObjectAndPrintsSuccessSummary(t *testing.T) {
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
	if len(repo.getRecipeCalls) != 0 {
		t.Fatalf("default create should not re-fetch recipe; got = %v", repo.getRecipeCalls)
	}
	if got := stdout.String(); got != "Successfully created recipe created-id\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRun_CreateWithJSONFlagPrintsCreatedRecipe(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, stdout, _ := testRunner(`{"name":"Pancakes"}`, repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"create", "-", "--json"}); err != nil {
		t.Fatalf("Run create --json: %v", err)
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

func TestRun_PatchMergesProvidedFieldsAndPrintsSuccessSummary(t *testing.T) {
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

	got := stdout.String()
	if !strings.HasPrefix(got, "Successfully updated recipe recipe-1") {
		t.Fatalf("stdout = %q, want success summary", got)
	}
	if !strings.Contains(got, "name") || !strings.Contains(got, "ingredients") {
		t.Fatalf("stdout = %q, want changed-field summary", got)
	}
}

func TestRun_PatchWithJSONFlagPrintsUpdatedRecipe(t *testing.T) {
	repo := &fakeRepo{
		recipes: []types.Recipe{{ID: "recipe-1", Name: "Old", Description: "Keep"}},
	}
	var factoryCalls int
	r, stdout, _ := testRunner(`{"name":"New"}`, repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"patch", "recipe-1", "-", "--json"}); err != nil {
		t.Fatalf("Run patch --json: %v", err)
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

func TestRun_AddPhotoReadsFileAndPrintsSuccessSummary(t *testing.T) {
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
	if len(repo.getRecipeCalls) != 0 {
		t.Fatalf("default add-photo should not re-fetch recipe; got = %v", repo.getRecipeCalls)
	}
	if got := stdout.String(); got != "Successfully added photo photo-id to recipe recipe-1 (featured)\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRun_AddPhotoReadsBase64FromStdin(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, stdout, _ := testRunner(" aW1n\n", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"add-photo", " recipe-1 ", "-"}); err != nil {
		t.Fatalf("Run add-photo: %v", err)
	}
	if repo.addPhotoCalls != 1 {
		t.Fatalf("add photo calls = %d, want 1", repo.addPhotoCalls)
	}
	if repo.addedPhoto.ImageBase64 != "aW1n" || repo.addedPhoto.Featured {
		t.Fatalf("added photo = %#v", repo.addedPhoto)
	}
	if got := stdout.String(); got != "Successfully added photo photo-id to recipe recipe-1\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRun_AddPhotoWithJSONFlagPrintsUpdatedRecipeWithoutImageContents(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, stdout, _ := testRunner(" aW1n\n", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"add-photo", "recipe-1", "-", "--featured", "--json"}); err != nil {
		t.Fatalf("Run add-photo --json: %v", err)
	}
	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if out.ID != "recipe-1" || len(out.Photos) != 1 {
		t.Fatalf("output recipe = %#v", out)
	}
	if out.Photos[0].ImageBase64 != "" {
		t.Fatalf("photo image_base64 should be stripped; got = %q", out.Photos[0].ImageBase64)
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

func TestRun_DeletePhotoRemovesPhotoAndPrintsSuccessSummary(t *testing.T) {
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
	if len(repo.getRecipeCalls) != 0 {
		t.Fatalf("default delete-photo should not re-fetch recipe; got = %v", repo.getRecipeCalls)
	}
	if got := stdout.String(); got != "Successfully deleted photo photo-1 from recipe recipe-1\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRun_DeletePhotoWithJSONFlagPrintsUpdatedRecipe(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, stdout, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"delete-photo", "recipe-1", "photo-1", "--json"}); err != nil {
		t.Fatalf("Run delete-photo --json: %v", err)
	}
	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if out.ID != "recipe-1" || len(out.Photos) != 0 {
		t.Fatalf("output recipe = %#v", out)
	}
}

func TestRun_DeleteRemovesRecipeAndPrintsSuccessSummary(t *testing.T) {
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
	if got := stdout.String(); got != "Successfully deleted recipe recipe-1\n" {
		t.Fatalf("stdout = %q", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRun_SetFeaturedPhotoMarksPhotoAndPrintsSuccessSummary(t *testing.T) {
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
	if len(repo.getRecipeCalls) != 0 {
		t.Fatalf("default set-featured-photo should not re-fetch recipe; got = %v", repo.getRecipeCalls)
	}
	if got := stdout.String(); got != "Successfully set photo photo-1 as featured on recipe recipe-1\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRun_SetFeaturedPhotoWithJSONFlagPrintsUpdatedRecipeWithoutImageContents(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, stdout, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"set-featured-photo", "recipe-1", "photo-1", "--json"}); err != nil {
		t.Fatalf("Run set-featured-photo --json: %v", err)
	}
	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if out.ID != "recipe-1" || len(out.Photos) != 1 || !out.Photos[0].Featured {
		t.Fatalf("output recipe = %#v", out)
	}
	if out.Photos[0].ImageBase64 != "" {
		t.Fatalf("photo image_base64 should be stripped; got = %q", out.Photos[0].ImageBase64)
	}
}

func TestRun_ImportReadsJSONLinesAndPrintsSuccessSummary(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, stdout, _ := testRunner("{\"id\":\"1\",\"name\":\"One\"}\n\n{\"id\":\"2\",\"name\":\"Two\"}\n", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"import", "-"}); err != nil {
		t.Fatalf("Run import: %v", err)
	}
	if len(repo.importedRecipes) != 2 {
		t.Fatalf("imported recipes = %d, want 2", len(repo.importedRecipes))
	}
	if repo.importedRecipes[0].ID != "1" || repo.importedRecipes[1].ID != "2" {
		t.Fatalf("imported recipes = %#v", repo.importedRecipes)
	}
	if got := stdout.String(); got != "Successfully imported 2 recipes\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRun_ExportStripsPhotoContentsByDefault(t *testing.T) {
	repo := &fakeRepo{
		recipes: []types.Recipe{
			{
				ID:   "recipe-1",
				Name: "One",
				Photos: []types.Photo{
					{ID: "p1", ImageBase64: "aGVsbG8=", Featured: true},
					{ID: "p2", ImageBase64: "d29ybGQ=", Featured: false},
				},
			},
		},
	}
	var factoryCalls int
	r, stdout, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"export", "recipe-1"}); err != nil {
		t.Fatalf("Run export: %v", err)
	}
	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if len(out.Photos) != 2 {
		t.Fatalf("photos len = %d, want 2", len(out.Photos))
	}
	for i, photo := range out.Photos {
		if photo.ID == "" {
			t.Fatalf("photo %d id was stripped: %#v", i, photo)
		}
		if photo.ImageBase64 != "" {
			t.Fatalf("photo %d image_base64 = %q, want empty", i, photo.ImageBase64)
		}
	}
	if !out.Photos[0].Featured {
		t.Fatalf("photo 0 featured flag was lost")
	}
}

func TestRun_ExportKeepsPhotoContentsWhenFlagGiven(t *testing.T) {
	repo := &fakeRepo{
		recipes: []types.Recipe{
			{
				ID:   "recipe-1",
				Name: "One",
				Photos: []types.Photo{
					{ID: "p1", ImageBase64: "aGVsbG8=", Featured: true},
				},
			},
		},
	}
	var factoryCalls int
	r, stdout, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"export", "recipe-1", "--image-contents"}); err != nil {
		t.Fatalf("Run export --image-contents: %v", err)
	}
	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if len(out.Photos) != 1 || out.Photos[0].ImageBase64 != "aGVsbG8=" {
		t.Fatalf("photo contents not preserved: %#v", out.Photos)
	}
}

func TestRun_ExportRejectsUnknownFlag(t *testing.T) {
	repo := &fakeRepo{recipes: []types.Recipe{{ID: "recipe-1"}}}
	var factoryCalls int
	r, _, stderr := testRunner("", repo, &factoryCalls)

	err := r.Run(context.Background(), []string{"export", "recipe-1", "--bogus"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "--image-contents") {
		t.Fatalf("stderr = %q, want usage mentioning --image-contents", stderr.String())
	}
}

func TestRun_ExportAllStripsPhotoContentsByDefault(t *testing.T) {
	repo := &fakeRepo{
		recipes: []types.Recipe{
			{ID: "1", Name: "One", Photos: []types.Photo{{ID: "p1", ImageBase64: "aGVsbG8="}}},
			{ID: "2", Name: "Two", Photos: []types.Photo{{ID: "p2", ImageBase64: "d29ybGQ="}}},
		},
	}
	var factoryCalls int
	r, stdout, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"export-all"}); err != nil {
		t.Fatalf("Run export-all: %v", err)
	}
	lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("export-all wrote %d lines, want 2: %q", len(lines), stdout.String())
	}
	for i, line := range lines {
		var out types.Recipe
		if err := json.Unmarshal([]byte(line), &out); err != nil {
			t.Fatalf("line %d not JSON: %v\n%s", i, err, line)
		}
		if len(out.Photos) != 1 || out.Photos[0].ID == "" || out.Photos[0].ImageBase64 != "" {
			t.Fatalf("line %d photo not stripped: %#v", i, out.Photos)
		}
	}
}

func TestRun_ExportAllKeepsPhotoContentsWhenFlagGiven(t *testing.T) {
	repo := &fakeRepo{
		recipes: []types.Recipe{
			{ID: "1", Name: "One", Photos: []types.Photo{{ID: "p1", ImageBase64: "aGVsbG8="}}},
		},
	}
	var factoryCalls int
	r, stdout, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"export-all", "--image-contents"}); err != nil {
		t.Fatalf("Run export-all --image-contents: %v", err)
	}
	var out types.Recipe
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout.String())), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if len(out.Photos) != 1 || out.Photos[0].ImageBase64 != "aGVsbG8=" {
		t.Fatalf("photo contents not preserved: %#v", out.Photos)
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
