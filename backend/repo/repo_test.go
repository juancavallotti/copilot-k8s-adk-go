package repo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	recipeops "juancavallotti.com/recipes-repo/internal/dbops/recipes"
	skillops "juancavallotti.com/recipes-repo/internal/dbops/skills"
	traceops "juancavallotti.com/recipes-repo/internal/dbops/traces"
	recipesvc "juancavallotti.com/recipes-repo/internal/service/recipes"
	skillsvc "juancavallotti.com/recipes-repo/internal/service/skills"
	tracesvc "juancavallotti.com/recipes-repo/internal/service/traces"
)

const testUUID = "550e8400-e29b-41d4-a716-446655440000"

func newMockRepo(t *testing.T) (*Repo, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	r := &Repo{
		recipes: recipesvc.NewService(recipeops.NewStore(db)),
		traces:  tracesvc.NewService(traceops.NewStore(db)),
		skills:  skillsvc.NewService(skillops.NewStore(db)),
		pool:    db,
	}
	return r, mock, func() { db.Close() }
}

// TestRepo_RecipeMethodsDelegateToRecipeService verifies that a recipe-flavored
// Repo method (GetRecipe) reaches the underlying store via the service. The
// service-layer tests cover the full method surface; this is a wiring smoke
// test.
func TestRepo_RecipeMethodsDelegateToRecipeService(t *testing.T) {
	t.Parallel()
	r, mock, cleanup := newMockRepo(t)
	defer cleanup()

	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM recipes WHERE id").
		WithArgs(testUUID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "image", "created_at", "updated_at"}).
			AddRow(testUUID, "from-recipe", "", "", "", ts, ts))
	mock.ExpectQuery("FROM recipes_ingredients").
		WithArgs(testUUID).
		WillReturnRows(sqlmock.NewRows([]string{"quantity", "unit", "name"}))
	mock.ExpectQuery("FROM steps").
		WithArgs(testUUID).
		WillReturnRows(sqlmock.NewRows([]string{"instruction"}))
	mock.ExpectQuery("FROM recipes_images").
		WithArgs(testUUID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "image_base64", "is_featured"}))

	got, err := r.GetRecipe(context.Background(), testUUID)
	if err != nil {
		t.Fatalf("GetRecipe: %v", err)
	}
	if got.ID != testUUID || got.Name != "from-recipe" {
		t.Fatalf("got = %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// TestRepo_TraceMethodsDelegateToTraceService verifies a trace-flavored Repo
// method (DeleteEvent) reaches the underlying store.
func TestRepo_TraceMethodsDelegateToTraceService(t *testing.T) {
	t.Parallel()
	r, mock, cleanup := newMockRepo(t)
	defer cleanup()

	ts := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	raw := json.RawMessage(`{"msg":"ok"}`)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO events").
		WithArgs("event-1", ts, "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO traces").
		WithArgs("event-1", ts, []byte(raw)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := r.LogTrace(context.Background(), "event-1", ts, raw); err != nil {
		t.Fatalf("LogTrace: %v", err)
	}

	mock.ExpectExec("DELETE FROM events").
		WithArgs("event-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := r.DeleteEvent(context.Background(), "event-1"); err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// TestRepo_SkillMethodsDelegateToSkillService verifies a skill-flavored Repo
// method (GetSkill) reaches the underlying store.
func TestRepo_SkillMethodsDelegateToSkillService(t *testing.T) {
	t.Parallel()
	r, mock, cleanup := newMockRepo(t)
	defer cleanup()

	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM skills").
		WithArgs(testUUID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "content", "created_at", "updated_at"}).
			AddRow(testUUID, "prep", "Prepare ingredients", "Instructions", now, now))

	got, err := r.GetSkill(context.Background(), testUUID)
	if err != nil {
		t.Fatalf("GetSkill: %v", err)
	}
	if got.ID != testUUID || got.Name != "prep" {
		t.Fatalf("got = %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
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
