package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	types "juancavallotti.com/recipe-types"
	repo "juancavallotti.com/recipes-repo"
)

// fakeRepo is the in-memory stub used by command tests. It satisfies each
// narrow domain interface (RecipeRepo/TraceRepo/SkillRepo/EmbedRepo) so
// tests can pass it straight to cmd methods. There is no runner-level
// union: r.Run(...) goes through openRepo (a package var defaulting to
// repo.NewRepo) for production, and tests of cmd behavior bypass Run by
// calling r.cmdXxx(...) directly with this fake.
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

	logTraceCalls   int
	logTraceEntries []traceEntry
	logTraceErr     error

	listEventsCalls  int
	listEventsLimit  int
	listEventsOffset int
	listEventsResult []types.Event
	listEventsErr    error

	listTracesCalls   int
	listTracesEventID string
	listTracesLimit   int
	listTracesOffset  int
	listTracesResult  []types.Trace
	listTracesErr     error

	listSkillsCalls  int
	listSkillsResult []types.Skill
	listSkillsErr    error

	getSkillByNameCalls  int
	getSkillByNameArg    string
	getSkillByNameResult types.Skill
	getSkillByNameErr    error

	embedCalls  int
	embedInput  string
	embedResult []float32
	embedErr    error

	reindexCalls   int
	reindexOpts    repo.ReindexOptions
	reindexReports []repo.IndexRecipeReport
	reindexErr     error

	reindexEventsCalls   int
	reindexEventsOpts    repo.ReindexEventsOptions
	reindexEventsReports []repo.IndexEventReport
	reindexEventsErr     error

	searchRecipesCalls       int
	searchRecipesQuery       string
	searchRecipesLimit       int
	searchRecipesResult      []types.RecipeMatch
	searchRecipesErr         error
	searchRecipeChunksCalls  int
	searchRecipeChunksQuery  string
	searchRecipeChunksLimit  int
	searchRecipeChunksResult []types.RecipeHit
	searchRecipeChunksErr    error
	searchEventsCalls        int
	searchEventsQuery        string
	searchEventsLimit        int
	searchEventsResult       []types.EventMatch
	searchEventsErr          error
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

func (f *fakeRepo) ListSkills(ctx context.Context) ([]types.Skill, error) {
	f.listSkillsCalls++
	return f.listSkillsResult, f.listSkillsErr
}

func (f *fakeRepo) GetSkillByName(ctx context.Context, name string) (types.Skill, error) {
	f.getSkillByNameCalls++
	f.getSkillByNameArg = name
	return f.getSkillByNameResult, f.getSkillByNameErr
}

func (f *fakeRepo) Embed(ctx context.Context, text string) ([]float32, error) {
	f.embedCalls++
	f.embedInput = text
	return f.embedResult, f.embedErr
}

func (f *fakeRepo) ReindexRecipes(ctx context.Context, opts repo.ReindexOptions) error {
	f.reindexCalls++
	f.reindexOpts = opts
	for _, rep := range f.reindexReports {
		if opts.OnReport != nil {
			opts.OnReport(rep)
		}
	}
	return f.reindexErr
}

func (f *fakeRepo) ReindexEvents(ctx context.Context, opts repo.ReindexEventsOptions) error {
	f.reindexEventsCalls++
	f.reindexEventsOpts = opts
	for _, rep := range f.reindexEventsReports {
		if opts.OnReport != nil {
			opts.OnReport(rep)
		}
	}
	return f.reindexEventsErr
}

func (f *fakeRepo) SearchRecipes(ctx context.Context, query string, limit int) ([]types.RecipeMatch, error) {
	f.searchRecipesCalls++
	f.searchRecipesQuery = query
	f.searchRecipesLimit = limit
	return f.searchRecipesResult, f.searchRecipesErr
}

func (f *fakeRepo) SearchRecipeChunks(ctx context.Context, query string, limit int) ([]types.RecipeHit, error) {
	f.searchRecipeChunksCalls++
	f.searchRecipeChunksQuery = query
	f.searchRecipeChunksLimit = limit
	return f.searchRecipeChunksResult, f.searchRecipeChunksErr
}

func (f *fakeRepo) SearchEvents(ctx context.Context, query string, limit int) ([]types.EventMatch, error) {
	f.searchEventsCalls++
	f.searchEventsQuery = query
	f.searchEventsLimit = limit
	return f.searchEventsResult, f.searchEventsErr
}

// testRunner builds a Runner with provided stdin and captured stdout/stderr.
// Use it alongside r.cmdXxx(...) calls for cmd-behavior tests; r.Run(...) is
// reserved for tests of the switch-level arg validation and help/schema
// paths. When r.Run does open a repo it goes through the package-level
// openRepo (repo.NewRepo by default), which is safe for arg-error tests
// because sql.Open is lazy and the noop embeddings client requires no DB.
func testRunner(stdin string) (Runner, *bytes.Buffer, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer
	r := Runner{
		stdin:  strings.NewReader(stdin),
		stdout: &stdout,
		stderr: &stderr,
		logger: slog.New(slog.NewJSONHandler(&stderr, nil)),
	}
	return r, &stdout, &stderr
}

func TestRun_NoArgsPrintsHelp(t *testing.T) {
	r, stdout, stderr := testRunner("")

	if err := r.Run(context.Background(), nil); err != nil {
		t.Fatalf("err = %v, want nil", err)
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
			r, stdout, stderr := testRunner("")

			if err := r.Run(context.Background(), []string{flag}); err != nil {
				t.Fatalf("err = %v, want nil", err)
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

func TestRun_SchemaPrintsValidJSON(t *testing.T) {
	r, stdout, stderr := testRunner("")

	if err := r.Run(context.Background(), []string{"schema"}); err != nil {
		t.Fatalf("Run schema: %v", err)
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

func TestCmdCreate_ReadsRecipeObjectAndPrintsSuccessSummary(t *testing.T) {
	repo := &fakeRepo{}
	r, stdout, _ := testRunner(`{
		"name": "Pancakes",
		"description": "Breakfast",
		"ingredients": ["1 cup flour"],
		"instructions": ["Mix"],
		"category": "breakfast",
		"image": "pancakes.jpg"
	}`)

	if err := r.cmdCreate(context.Background(), repo, "-", false); err != nil {
		t.Fatalf("cmdCreate: %v", err)
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

func TestCmdCreate_WithJSONFlagPrintsCreatedRecipe(t *testing.T) {
	repo := &fakeRepo{}
	r, stdout, _ := testRunner(`{"name":"Pancakes"}`)

	if err := r.cmdCreate(context.Background(), repo, "-", true); err != nil {
		t.Fatalf("cmdCreate --json: %v", err)
	}
	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if out.ID != "created-id" || out.Name != "Pancakes" {
		t.Fatalf("output recipe = %#v", out)
	}
}

func TestCmdCreate_RejectsUnknownFields(t *testing.T) {
	repo := &fakeRepo{}
	r, _, _ := testRunner(`{"id":"client-id","name":"Pancakes"}`)

	err := r.cmdCreate(context.Background(), repo, "-", false)
	if err == nil {
		t.Fatal("cmdCreate returned nil, want unknown field error")
	}
	if !strings.Contains(err.Error(), `unknown field "id"`) {
		t.Fatalf("err = %v, want unknown field id", err)
	}
	if repo.createRecipeCalls != 0 {
		t.Fatalf("create calls = %d, want 0", repo.createRecipeCalls)
	}
}

func TestCmdPatch_MergesProvidedFieldsAndPrintsSuccessSummary(t *testing.T) {
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
	r, stdout, _ := testRunner(`{"name":"New","ingredients":["2 cups flour"]}`)

	if err := r.cmdPatch(context.Background(), repo, " recipe-1 ", "-", false); err != nil {
		t.Fatalf("cmdPatch: %v", err)
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

func TestCmdPatch_WithJSONFlagPrintsUpdatedRecipe(t *testing.T) {
	repo := &fakeRepo{
		recipes: []types.Recipe{{ID: "recipe-1", Name: "Old", Description: "Keep"}},
	}
	r, stdout, _ := testRunner(`{"name":"New"}`)

	if err := r.cmdPatch(context.Background(), repo, "recipe-1", "-", true); err != nil {
		t.Fatalf("cmdPatch --json: %v", err)
	}
	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if out.Name != "New" || out.Description != "Keep" {
		t.Fatalf("output recipe = %#v", out)
	}
}

func TestCmdPatch_RejectsEmptyPatch(t *testing.T) {
	repo := &fakeRepo{}
	r, _, _ := testRunner(`{}`)

	err := r.cmdPatch(context.Background(), repo, "recipe-1", "-", false)
	if err == nil || err.Error() != "no fields to update" {
		t.Fatalf("err = %v, want no fields to update", err)
	}
	if repo.updateRecipeCalls != 0 {
		t.Fatalf("update calls = %d, want 0", repo.updateRecipeCalls)
	}
}

func TestCmdAddPhoto_ReadsFileAndPrintsSuccessSummary(t *testing.T) {
	repo := &fakeRepo{}
	tmp := t.TempDir() + "/photo.bin"
	if err := os.WriteFile(tmp, []byte("img"), 0o600); err != nil {
		t.Fatal(err)
	}
	r, stdout, _ := testRunner("")

	if err := r.cmdAddPhoto(context.Background(), repo, " recipe-1 ", tmp, true, false); err != nil {
		t.Fatalf("cmdAddPhoto: %v", err)
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

func TestCmdAddPhoto_ReadsBase64FromStdin(t *testing.T) {
	repo := &fakeRepo{}
	r, stdout, _ := testRunner(" aW1n\n")

	if err := r.cmdAddPhoto(context.Background(), repo, " recipe-1 ", "-", false, false); err != nil {
		t.Fatalf("cmdAddPhoto: %v", err)
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

func TestCmdAddPhoto_WithJSONFlagPrintsUpdatedRecipeWithoutImageContents(t *testing.T) {
	repo := &fakeRepo{}
	r, stdout, _ := testRunner(" aW1n\n")

	if err := r.cmdAddPhoto(context.Background(), repo, "recipe-1", "-", true, true); err != nil {
		t.Fatalf("cmdAddPhoto --json: %v", err)
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

func TestCmdAddPhoto_RejectsInvalidBase64FromStdin(t *testing.T) {
	repo := &fakeRepo{}
	r, _, _ := testRunner("not base64!")

	err := r.cmdAddPhoto(context.Background(), repo, "recipe-1", "-", false, false)
	if err == nil || !strings.Contains(err.Error(), "invalid base64 image data") {
		t.Fatalf("err = %v, want invalid base64 image data", err)
	}
	if repo.addPhotoCalls != 0 {
		t.Fatalf("add photo calls = %d, want 0", repo.addPhotoCalls)
	}
}

func TestCmdDeletePhoto_RemovesPhotoAndPrintsSuccessSummary(t *testing.T) {
	repo := &fakeRepo{}
	r, stdout, _ := testRunner("")

	if err := r.cmdDeletePhoto(context.Background(), repo, " recipe-1 ", " photo-1 ", false); err != nil {
		t.Fatalf("cmdDeletePhoto: %v", err)
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

func TestCmdDeletePhoto_WithJSONFlagPrintsUpdatedRecipe(t *testing.T) {
	repo := &fakeRepo{}
	r, stdout, _ := testRunner("")

	if err := r.cmdDeletePhoto(context.Background(), repo, "recipe-1", "photo-1", true); err != nil {
		t.Fatalf("cmdDeletePhoto --json: %v", err)
	}
	var out types.Recipe
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if out.ID != "recipe-1" || len(out.Photos) != 0 {
		t.Fatalf("output recipe = %#v", out)
	}
}

func TestCmdDelete_RemovesRecipeAndPrintsSuccessSummary(t *testing.T) {
	repo := &fakeRepo{}
	r, stdout, stderr := testRunner("")

	if err := r.cmdDelete(context.Background(), repo, " recipe-1 "); err != nil {
		t.Fatalf("cmdDelete: %v", err)
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

func TestCmdSetFeaturedPhoto_MarksPhotoAndPrintsSuccessSummary(t *testing.T) {
	repo := &fakeRepo{}
	r, stdout, _ := testRunner("")

	if err := r.cmdSetFeaturedPhoto(context.Background(), repo, " recipe-1 ", " photo-1 ", false); err != nil {
		t.Fatalf("cmdSetFeaturedPhoto: %v", err)
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

func TestCmdSetFeaturedPhoto_WithJSONFlagPrintsUpdatedRecipeWithoutImageContents(t *testing.T) {
	repo := &fakeRepo{}
	r, stdout, _ := testRunner("")

	if err := r.cmdSetFeaturedPhoto(context.Background(), repo, "recipe-1", "photo-1", true); err != nil {
		t.Fatalf("cmdSetFeaturedPhoto --json: %v", err)
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

func TestCmdImport_ReadsJSONLinesAndPrintsSuccessSummary(t *testing.T) {
	repo := &fakeRepo{}
	r, stdout, _ := testRunner("{\"id\":\"1\",\"name\":\"One\"}\n\n{\"id\":\"2\",\"name\":\"Two\"}\n")

	if err := r.cmdImport(context.Background(), repo, "-"); err != nil {
		t.Fatalf("cmdImport: %v", err)
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

func TestCmdExport_StripsPhotoContentsByDefault(t *testing.T) {
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
	r, stdout, _ := testRunner("")

	if err := r.cmdExport(context.Background(), repo, "recipe-1", false); err != nil {
		t.Fatalf("cmdExport: %v", err)
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

func TestCmdExport_KeepsPhotoContentsWhenFlagGiven(t *testing.T) {
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
	r, stdout, _ := testRunner("")

	if err := r.cmdExport(context.Background(), repo, "recipe-1", true); err != nil {
		t.Fatalf("cmdExport --image-contents: %v", err)
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
	r, _, stderr := testRunner("")

	err := r.Run(context.Background(), []string{"export", "recipe-1", "--bogus"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "--image-contents") {
		t.Fatalf("stderr = %q, want usage mentioning --image-contents", stderr.String())
	}
}

func TestCmdExportAll_StripsPhotoContentsByDefault(t *testing.T) {
	repo := &fakeRepo{
		recipes: []types.Recipe{
			{ID: "1", Name: "One", Photos: []types.Photo{{ID: "p1", ImageBase64: "aGVsbG8="}}},
			{ID: "2", Name: "Two", Photos: []types.Photo{{ID: "p2", ImageBase64: "d29ybGQ="}}},
		},
	}
	r, stdout, _ := testRunner("")

	if err := r.cmdExportAll(context.Background(), repo, false); err != nil {
		t.Fatalf("cmdExportAll: %v", err)
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

func TestCmdExportAll_KeepsPhotoContentsWhenFlagGiven(t *testing.T) {
	repo := &fakeRepo{
		recipes: []types.Recipe{
			{ID: "1", Name: "One", Photos: []types.Photo{{ID: "p1", ImageBase64: "aGVsbG8="}}},
		},
	}
	r, stdout, _ := testRunner("")

	if err := r.cmdExportAll(context.Background(), repo, true); err != nil {
		t.Fatalf("cmdExportAll --image-contents: %v", err)
	}
	var out types.Recipe
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout.String())), &out); err != nil {
		t.Fatalf("output is not recipe JSON: %v\n%s", err, stdout.String())
	}
	if len(out.Photos) != 1 || out.Photos[0].ImageBase64 != "aGVsbG8=" {
		t.Fatalf("photo contents not preserved: %#v", out.Photos)
	}
}

func TestCmdList_PrintsTable(t *testing.T) {
	repo := &fakeRepo{
		recipes: []types.Recipe{
			{ID: "1", Name: "One"},
			{ID: "2", Name: "Two"},
		},
	}
	r, stdout, _ := testRunner("")

	if err := r.cmdList(context.Background(), repo); err != nil {
		t.Fatalf("cmdList: %v", err)
	}
	got := stdout.String()
	for _, want := range []string{"ID", "TITLE", "1", "One", "2", "Two"} {
		if !strings.Contains(got, want) {
			t.Fatalf("list output = %q, want %q", got, want)
		}
	}
}

func TestCmdEmbedTest_PrintsDimensions(t *testing.T) {
	t.Parallel()
	repo := &fakeRepo{embedResult: []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6}}
	r, stdout, _ := testRunner("")
	if err := r.cmdEmbedTest(context.Background(), repo, "hello world"); err != nil {
		t.Fatalf("cmdEmbedTest: %v", err)
	}
	if repo.embedCalls != 1 || repo.embedInput != "hello world" {
		t.Fatalf("embed not called as expected: calls=%d input=%q", repo.embedCalls, repo.embedInput)
	}
	out := stdout.String()
	if !strings.Contains(out, "dimensions=6") {
		t.Fatalf("stdout = %q, want dimensions=6", out)
	}
}

func TestRun_EmbedTestMissingArgIsUsageError(t *testing.T) {
	t.Parallel()
	r, _, stderr := testRunner("")
	err := r.Run(context.Background(), []string{"embed-test"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "embed-test") {
		t.Fatalf("stderr = %q, want usage hint", stderr.String())
	}
}

func TestCmdReindex_RecipesJSON(t *testing.T) {
	t.Parallel()
	repo := &fakeRepo{
		reindexReports: []repo.IndexRecipeReport{
			{ID: "r1", Status: "ok"},
			{ID: "r2", Status: "error", Error: "embed boom"},
		},
	}
	r, stdout, _ := testRunner("")
	err := r.cmdReindex(context.Background(), repo, []string{"--target", "recipes", "--json"})
	if err == nil {
		t.Fatal("expected non-nil err (failed row)")
	}
	out := stdout.String()
	if !strings.Contains(out, `"id":"r1"`) || !strings.Contains(out, `"status":"ok"`) {
		t.Fatalf("stdout = %q, want r1 ok", out)
	}
	if !strings.Contains(out, `"status":"error"`) || !strings.Contains(out, "embed boom") {
		t.Fatalf("stdout = %q, want r2 error", out)
	}
}

func TestCmdReindex_RecipesHumanOutput(t *testing.T) {
	t.Parallel()
	repo := &fakeRepo{
		reindexReports: []repo.IndexRecipeReport{{ID: "r1", Status: "ok"}},
	}
	r, stdout, _ := testRunner("")
	if err := r.cmdReindex(context.Background(), repo, []string{"--target", "recipes", "--force", "--limit", "10"}); err != nil {
		t.Fatalf("cmdReindex: %v", err)
	}
	if !repo.reindexOpts.Force || repo.reindexOpts.Limit != 10 {
		t.Fatalf("opts = %+v, want force=true limit=10", repo.reindexOpts)
	}
	out := stdout.String()
	if !strings.Contains(out, "r1\tok") {
		t.Fatalf("stdout = %q, want r1 ok", out)
	}
	if !strings.Contains(out, "indexed 1 recipe(s)") {
		t.Fatalf("stdout = %q, want summary line", out)
	}
}

func TestCmdReindex_EventsJSON(t *testing.T) {
	t.Parallel()
	repo := &fakeRepo{
		reindexEventsReports: []repo.IndexEventReport{
			{ID: "inv-1", Status: "ok"},
		},
	}
	r, stdout, _ := testRunner("")
	if err := r.cmdReindex(context.Background(), repo, []string{"--target", "events", "--json"}); err != nil {
		t.Fatalf("cmdReindex events: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, `"id":"inv-1"`) || !strings.Contains(out, `"status":"ok"`) {
		t.Fatalf("stdout = %q, want inv-1 ok", out)
	}
}

func TestCmdReindex_AllInvokesBothTargets(t *testing.T) {
	t.Parallel()
	repo := &fakeRepo{
		reindexReports:       []repo.IndexRecipeReport{{ID: "r1", Status: "ok"}},
		reindexEventsReports: []repo.IndexEventReport{{ID: "inv-1", Status: "ok"}},
	}
	r, stdout, _ := testRunner("")
	if err := r.cmdReindex(context.Background(), repo, []string{"--target", "all"}); err != nil {
		t.Fatalf("cmdReindex all: %v", err)
	}
	if repo.reindexCalls != 1 || repo.reindexEventsCalls != 1 {
		t.Fatalf("calls: recipes=%d events=%d, want 1 each", repo.reindexCalls, repo.reindexEventsCalls)
	}
	out := stdout.String()
	if !strings.Contains(out, "r1\tok") || !strings.Contains(out, "inv-1\tok") {
		t.Fatalf("stdout = %q, want both targets reported", out)
	}
}

func TestCmdSearch_RecipesHumanOutput(t *testing.T) {
	t.Parallel()
	repo := &fakeRepo{
		searchRecipeChunksResult: []types.RecipeHit{
			{ID: "r1", Name: "Carbonara", Chunk: "guanciale, pecorino, eggs", Score: 0.91},
			{ID: "r2", Name: "Pesto", Chunk: "basil, pine nuts", Score: 0.72},
		},
	}
	r, stdout, _ := testRunner("")
	if err := r.cmdSearch(context.Background(), repo, "recipes", []string{"creamy", "pasta"}); err != nil {
		t.Fatalf("cmdSearch: %v", err)
	}
	if repo.searchRecipeChunksQuery != "creamy pasta" || repo.searchRecipeChunksLimit != 10 {
		t.Fatalf("query=%q limit=%d, want \"creamy pasta\" / 10", repo.searchRecipeChunksQuery, repo.searchRecipeChunksLimit)
	}
	out := stdout.String()
	for _, want := range []string{"SCORE", "ID", "TITLE", "CHUNK", "0.9100", "r1", "Carbonara", "guanciale", "0.7200", "r2", "Pesto", "basil"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout = %q, missing %q", out, want)
		}
	}
}

func TestCmdSearch_RecipesJSON(t *testing.T) {
	t.Parallel()
	repo := &fakeRepo{
		searchRecipeChunksResult: []types.RecipeHit{
			{ID: "r1", Name: "Carbonara", Chunk: "eggs, pecorino", Score: 0.91},
		},
	}
	r, stdout, _ := testRunner("")
	if err := r.cmdSearch(context.Background(), repo, "recipes", []string{"pasta", "--limit", "5", "--json"}); err != nil {
		t.Fatalf("cmdSearch --json: %v", err)
	}
	if repo.searchRecipeChunksLimit != 5 {
		t.Fatalf("limit = %d, want 5", repo.searchRecipeChunksLimit)
	}
	out := stdout.String()
	for _, want := range []string{`"id":"r1"`, `"name":"Carbonara"`, `"chunk":"eggs, pecorino"`, `"score":0.91`} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout = %q, missing %q", out, want)
		}
	}
	if strings.Contains(out, "image_base64") || strings.Contains(out, "photos") {
		t.Fatalf("stdout = %q, slim JSON should not contain recipe photos", out)
	}
}

func TestCmdSearch_RecipesDisabledReportsToStderr(t *testing.T) {
	t.Parallel()
	repoFake := &fakeRepo{searchRecipeChunksErr: repo.ErrSearchDisabled}
	r, _, stderr := testRunner("")
	err := r.cmdSearch(context.Background(), repoFake, "recipes", []string{"pasta"})
	if !errors.Is(err, repo.ErrSearchDisabled) {
		t.Fatalf("err = %v, want ErrSearchDisabled", err)
	}
	if !strings.Contains(stderr.String(), "search disabled") {
		t.Fatalf("stderr = %q, want hint about disabled search", stderr.String())
	}
}

func TestCmdSearch_EventsHumanOutput(t *testing.T) {
	t.Parallel()
	repo := &fakeRepo{
		searchEventsResult: []types.EventMatch{
			{Event: types.Event{EventID: "inv-1", UserPrompt: "find recipes with chicken"}, Score: 0.8},
		},
	}
	r, stdout, _ := testRunner("")
	if err := r.cmdSearch(context.Background(), repo, "events", []string{"chicken"}); err != nil {
		t.Fatalf("cmdSearch events: %v", err)
	}
	out := stdout.String()
	for _, want := range []string{"SCORE", "EVENT_ID", "PROMPT", "0.8000", "inv-1", "find recipes with chicken"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout = %q, missing %q", out, want)
		}
	}
}

func TestCmdSearch_MissingQueryIsUsageError(t *testing.T) {
	t.Parallel()
	repo := &fakeRepo{}
	r, _, stderr := testRunner("")
	err := r.cmdSearch(context.Background(), repo, "recipes", []string{})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "search-recipes") {
		t.Fatalf("stderr = %q, want usage hint", stderr.String())
	}
}

func TestCmdReindex_MissingTargetIsUsageError(t *testing.T) {
	t.Parallel()
	repo := &fakeRepo{}
	r, _, stderr := testRunner("")
	err := r.cmdReindex(context.Background(), repo, []string{})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "--target") {
		t.Fatalf("stderr = %q, want usage hint", stderr.String())
	}
}
