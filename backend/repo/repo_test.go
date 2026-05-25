package repo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	types "juancavallotti.com/recipe-types"
	recipeops "juancavallotti.com/recipes-repo/internal/dbops/recipes"
	skillops "juancavallotti.com/recipes-repo/internal/dbops/skills"
	traceops "juancavallotti.com/recipes-repo/internal/dbops/traces"
	recipesvc "juancavallotti.com/recipes-repo/internal/service/recipes"
	skillsvc "juancavallotti.com/recipes-repo/internal/service/skills"
	tracesvc "juancavallotti.com/recipes-repo/internal/service/traces"
)

const testUUID = "550e8400-e29b-41d4-a716-446655440000"

type fakeRecipeStore struct {
	getRecipesCalls       int
	getRecipeID           string
	createRecipe          types.Recipe
	createWithIDRecipe    types.Recipe
	updateRecipe          types.Recipe
	addPhotoRecipeID      string
	addPhoto              types.Photo
	deletePhotoRecipeID   string
	deletePhotoID         string
	setFeaturedRecipeID   string
	setFeaturedPhotoID    string
	deleteRecipeID        string
	importGetRecipeCalls  int
	importCreateWithIDErr error
}

func (f *fakeRecipeStore) GetRecipes(ctx context.Context) ([]types.Recipe, error) {
	f.getRecipesCalls++
	return []types.Recipe{{ID: testUUID, Name: "from-recipes"}}, nil
}

func (f *fakeRecipeStore) GetRecipe(ctx context.Context, id string) (types.Recipe, error) {
	f.getRecipeID = id
	f.importGetRecipeCalls++
	return types.Recipe{ID: id, Name: "from-recipe"}, nil
}

func (f *fakeRecipeStore) CreateRecipe(ctx context.Context, recipe types.Recipe) (string, error) {
	f.createRecipe = recipe
	return testUUID, nil
}

func (f *fakeRecipeStore) CreateRecipeWithID(ctx context.Context, recipe types.Recipe) error {
	f.createWithIDRecipe = recipe
	return f.importCreateWithIDErr
}

func (f *fakeRecipeStore) UpdateRecipe(ctx context.Context, recipe types.Recipe) error {
	f.updateRecipe = recipe
	return nil
}

func (f *fakeRecipeStore) AddRecipePhoto(ctx context.Context, recipeID string, photo types.Photo) (string, error) {
	f.addPhotoRecipeID = recipeID
	f.addPhoto = photo
	return "photo-id", nil
}

func (f *fakeRecipeStore) DeleteRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	f.deletePhotoRecipeID = recipeID
	f.deletePhotoID = photoID
	return nil
}

func (f *fakeRecipeStore) SetFeaturedRecipePhoto(ctx context.Context, recipeID string, photoID string) error {
	f.setFeaturedRecipeID = recipeID
	f.setFeaturedPhotoID = photoID
	return nil
}

func (f *fakeRecipeStore) DeleteRecipe(ctx context.Context, id string) error {
	f.deleteRecipeID = id
	return nil
}

func (f *fakeRecipeStore) IndexRecipe(ctx context.Context, id string) error {
	return nil
}

func (f *fakeRecipeStore) ReindexRecipes(ctx context.Context, opts recipeops.ReindexOptions) error {
	return nil
}

type fakeTraceStore struct {
	insertEventID      string
	insertOccurredAt   time.Time
	insertData         json.RawMessage
	listEventsLimit    int
	listEventsOffset   int
	listTracesEventID  string
	listTracesLimit    int
	listTracesOffset   int
	deleteAllCalls     int
	deleteEventByIDArg string
}

func (f *fakeTraceStore) InsertTrace(ctx context.Context, eventID string, occurredAt time.Time, data json.RawMessage) error {
	f.insertEventID = eventID
	f.insertOccurredAt = occurredAt
	f.insertData = data
	return nil
}

func (f *fakeTraceStore) ListEvents(ctx context.Context, limit, offset int) ([]types.Event, error) {
	f.listEventsLimit = limit
	f.listEventsOffset = offset
	return []types.Event{{EventID: "event-1"}}, nil
}

func (f *fakeTraceStore) ListTracesByEvent(ctx context.Context, eventID string, limit, offset int) ([]types.Trace, error) {
	f.listTracesEventID = eventID
	f.listTracesLimit = limit
	f.listTracesOffset = offset
	return []types.Trace{{ID: testUUID, EventID: eventID}}, nil
}

func (f *fakeTraceStore) DeleteAllEvents(ctx context.Context) error {
	f.deleteAllCalls++
	return nil
}

func (f *fakeTraceStore) DeleteEventByID(ctx context.Context, eventID string) error {
	f.deleteEventByIDArg = eventID
	return nil
}

type fakeSkillStore struct {
	getSkillID          string
	getSkillByNameArg   string
	createName          string
	createDescription   string
	createContent       string
	updateID            string
	updateDescription   string
	updateContent       string
	deleteID            string
	listSkillsCallCount int
}

func (f *fakeSkillStore) ListSkills(ctx context.Context) ([]types.Skill, error) {
	f.listSkillsCallCount++
	return []types.Skill{{ID: testUUID, Name: "prep"}}, nil
}

func (f *fakeSkillStore) GetSkill(ctx context.Context, id string) (types.Skill, error) {
	f.getSkillID = id
	return types.Skill{ID: id, Name: "prep"}, nil
}

func (f *fakeSkillStore) GetSkillByName(ctx context.Context, name string) (types.Skill, error) {
	f.getSkillByNameArg = name
	return types.Skill{ID: testUUID, Name: name}, nil
}

func (f *fakeSkillStore) CreateSkill(ctx context.Context, name, description, content string) (string, error) {
	f.createName = name
	f.createDescription = description
	f.createContent = content
	return testUUID, nil
}

func (f *fakeSkillStore) UpdateSkill(ctx context.Context, id, description, content string) error {
	f.updateID = id
	f.updateDescription = description
	f.updateContent = content
	return nil
}

func (f *fakeSkillStore) DeleteSkill(ctx context.Context, id string) error {
	f.deleteID = id
	return nil
}

func newTestRepo(recipeStore *fakeRecipeStore, traceStore *fakeTraceStore, skillStore *fakeSkillStore) *Repo {
	return &Repo{
		recipes: recipesvc.NewService(recipeStore),
		traces:  tracesvc.NewService(traceStore),
		skills:  skillsvc.NewService(skillStore),
	}
}

func TestRepo_RecipeMethodsDelegateToRecipeService(t *testing.T) {
	t.Parallel()
	store := &fakeRecipeStore{}
	r := newTestRepo(store, &fakeTraceStore{}, &fakeSkillStore{})
	ctx := context.Background()
	recipe := types.Recipe{Name: "Soup", Ingredients: []string{"water"}, Instructions: []string{"boil"}}
	recipeWithID := recipe
	recipeWithID.ID = testUUID
	photo := types.Photo{ImageBase64: "aW1n", Featured: true}

	if got, err := r.GetRecipes(ctx); err != nil || len(got) != 1 {
		t.Fatalf("GetRecipes = (%#v, %v), want one recipe and nil error", got, err)
	}
	if _, err := r.GetRecipe(ctx, testUUID); err != nil {
		t.Fatalf("GetRecipe: %v", err)
	}
	if _, err := r.CreateRecipe(ctx, recipe); err != nil {
		t.Fatalf("CreateRecipe: %v", err)
	}
	if err := r.UpdateRecipe(ctx, recipeWithID); err != nil {
		t.Fatalf("UpdateRecipe: %v", err)
	}
	if _, err := r.AddRecipePhoto(ctx, testUUID, photo); err != nil {
		t.Fatalf("AddRecipePhoto: %v", err)
	}
	if err := r.DeleteRecipePhoto(ctx, testUUID, testUUID); err != nil {
		t.Fatalf("DeleteRecipePhoto: %v", err)
	}
	if err := r.SetFeaturedRecipePhoto(ctx, testUUID, testUUID); err != nil {
		t.Fatalf("SetFeaturedRecipePhoto: %v", err)
	}
	if err := r.DeleteRecipe(ctx, testUUID); err != nil {
		t.Fatalf("DeleteRecipe: %v", err)
	}
	if err := r.ImportRecipe(ctx, recipeWithID); err != nil {
		t.Fatalf("ImportRecipe: %v", err)
	}

	if store.getRecipesCalls != 1 || store.getRecipeID != testUUID {
		t.Fatalf("read calls not delegated: getRecipes=%d getRecipeID=%q", store.getRecipesCalls, store.getRecipeID)
	}
	if store.createRecipe.Name != recipe.Name || store.updateRecipe.ID != testUUID {
		t.Fatalf("write calls not delegated: create=%#v update=%#v", store.createRecipe, store.updateRecipe)
	}
	if store.addPhotoRecipeID != testUUID || store.addPhoto.ImageBase64 != photo.ImageBase64 {
		t.Fatalf("photo add not delegated: recipeID=%q photo=%#v", store.addPhotoRecipeID, store.addPhoto)
	}
	if store.deletePhotoRecipeID != testUUID || store.deletePhotoID != testUUID {
		t.Fatalf("photo delete not delegated: recipeID=%q photoID=%q", store.deletePhotoRecipeID, store.deletePhotoID)
	}
	if store.setFeaturedRecipeID != testUUID || store.setFeaturedPhotoID != testUUID {
		t.Fatalf("set featured not delegated: recipeID=%q photoID=%q", store.setFeaturedRecipeID, store.setFeaturedPhotoID)
	}
	if store.deleteRecipeID != testUUID {
		t.Fatalf("delete not delegated: id=%q", store.deleteRecipeID)
	}
}

func TestRepo_TraceMethodsDelegateToTraceService(t *testing.T) {
	t.Parallel()
	store := &fakeTraceStore{}
	r := newTestRepo(&fakeRecipeStore{}, store, &fakeSkillStore{})
	ctx := context.Background()
	ts := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	raw := json.RawMessage(`{"msg":"ok"}`)

	if err := r.LogTrace(ctx, "event-1", ts, raw); err != nil {
		t.Fatalf("LogTrace: %v", err)
	}
	if got, err := r.ListEvents(ctx, 25, 5); err != nil || len(got) != 1 {
		t.Fatalf("ListEvents = (%#v, %v), want one event and nil error", got, err)
	}
	if got, err := r.ListTracesByEvent(ctx, "event-1", 10, 2); err != nil || len(got) != 1 {
		t.Fatalf("ListTracesByEvent = (%#v, %v), want one trace and nil error", got, err)
	}
	if err := r.DeleteAllEvents(ctx); err != nil {
		t.Fatalf("DeleteAllEvents: %v", err)
	}
	if err := r.DeleteEvent(ctx, "event-1"); err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}

	if store.insertEventID != "event-1" || !store.insertOccurredAt.Equal(ts) || string(store.insertData) != string(raw) {
		t.Fatalf("insert not delegated: eventID=%q occurredAt=%v data=%s", store.insertEventID, store.insertOccurredAt, store.insertData)
	}
	if store.listEventsLimit != 25 || store.listEventsOffset != 5 {
		t.Fatalf("list events paging = (%d, %d), want (25, 5)", store.listEventsLimit, store.listEventsOffset)
	}
	if store.listTracesEventID != "event-1" || store.listTracesLimit != 10 || store.listTracesOffset != 2 {
		t.Fatalf("list traces args = (%q, %d, %d), want (event-1, 10, 2)", store.listTracesEventID, store.listTracesLimit, store.listTracesOffset)
	}
	if store.deleteAllCalls != 1 || store.deleteEventByIDArg != "event-1" {
		t.Fatalf("delete calls not delegated: all=%d eventID=%q", store.deleteAllCalls, store.deleteEventByIDArg)
	}
}

func TestRepo_SkillMethodsDelegateToSkillService(t *testing.T) {
	t.Parallel()
	store := &fakeSkillStore{}
	r := newTestRepo(&fakeRecipeStore{}, &fakeTraceStore{}, store)
	ctx := context.Background()

	if got, err := r.ListSkills(ctx); err != nil || len(got) != 1 {
		t.Fatalf("ListSkills = (%#v, %v), want one skill and nil error", got, err)
	}
	if _, err := r.GetSkill(ctx, testUUID); err != nil {
		t.Fatalf("GetSkill: %v", err)
	}
	if _, err := r.GetSkillByName(ctx, "prep"); err != nil {
		t.Fatalf("GetSkillByName: %v", err)
	}
	if _, err := r.CreateSkill(ctx, "prep", "Prepare ingredients", "Instructions"); err != nil {
		t.Fatalf("CreateSkill: %v", err)
	}
	if err := r.UpdateSkill(ctx, testUUID, "Updated description", "Updated content"); err != nil {
		t.Fatalf("UpdateSkill: %v", err)
	}
	if err := r.DeleteSkill(ctx, testUUID); err != nil {
		t.Fatalf("DeleteSkill: %v", err)
	}

	if store.listSkillsCallCount != 1 || store.getSkillID != testUUID || store.getSkillByNameArg != "prep" {
		t.Fatalf("read calls not delegated: list=%d id=%q name=%q", store.listSkillsCallCount, store.getSkillID, store.getSkillByNameArg)
	}
	if store.createName != "prep" || store.createDescription != "Prepare ingredients" || store.createContent != "Instructions" {
		t.Fatalf("create not delegated: name=%q description=%q content=%q", store.createName, store.createDescription, store.createContent)
	}
	if store.updateID != testUUID || store.updateDescription != "Updated description" || store.updateContent != "Updated content" {
		t.Fatalf("update not delegated: id=%q description=%q content=%q", store.updateID, store.updateDescription, store.updateContent)
	}
	if store.deleteID != testUUID {
		t.Fatalf("delete not delegated: id=%q", store.deleteID)
	}
}

func TestRepo_ReExportedErrors(t *testing.T) {
	t.Parallel()
	if ErrRecipeNotFound != recipeops.ErrRecipeNotFound {
		t.Fatal("ErrRecipeNotFound does not re-export recipeops.ErrRecipeNotFound")
	}
	if ErrEventNotFound != traceops.ErrEventNotFound {
		t.Fatal("ErrEventNotFound does not re-export traceops.ErrEventNotFound")
	}
	if ErrSkillNotFound != skillops.ErrSkillNotFound {
		t.Fatal("ErrSkillNotFound does not re-export skillops.ErrSkillNotFound")
	}
}
