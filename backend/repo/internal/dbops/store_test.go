package dbops

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	types "juancavallotti.com/recipe-types"
)

func TestNewStore_nilPool(t *testing.T) {
	t.Parallel()
	s := NewStore(nil)
	if s == nil || s.db != nil {
		t.Fatalf("NewStore(nil) = %#v", s)
	}
}

func TestStore_nilDB_errors(t *testing.T) {
	t.Parallel()
	s := &Store{db: nil}
	ctx := context.Background()

	if _, err := s.GetRecipes(ctx); !errors.Is(err, errNilDB) {
		t.Fatalf("GetRecipes err = %v", err)
	}
	if _, err := s.GetRecipe(ctx, "550e8400-e29b-41d4-a716-446655440000"); !errors.Is(err, errNilDB) {
		t.Fatalf("GetRecipe err = %v", err)
	}
	if _, err := s.CreateRecipe(ctx, types.Recipe{}); !errors.Is(err, errNilDB) {
		t.Fatalf("CreateRecipe err = %v", err)
	}
	if err := s.UpdateRecipe(ctx, types.Recipe{ID: "550e8400-e29b-41d4-a716-446655440000"}); !errors.Is(err, errNilDB) {
		t.Fatalf("UpdateRecipe err = %v", err)
	}
	if err := s.DeleteRecipe(ctx, "550e8400-e29b-41d4-a716-446655440000"); !errors.Is(err, errNilDB) {
		t.Fatalf("DeleteRecipe err = %v", err)
	}
}

func TestStore_GetRecipe_invalidUUID(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	_, err = s.GetRecipe(context.Background(), "not-a-uuid")
	if err == nil || err.Error() == "" {
		t.Fatal("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_GetRecipes_contextCanceled(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = s.GetRecipes(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_GetRecipes_empty(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectQuery("SELECT id::text").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "image", "created_at", "updated_at"}))

	out, err := s.GetRecipes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("len = %d", len(out))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_GetRecipe_notFound(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	rid := "550e8400-e29b-41d4-a716-446655440000"
	mock.ExpectQuery("SELECT id::text, name, description, category, image, created_at, updated_at").
		WithArgs(rid).
		WillReturnError(sql.ErrNoRows)

	_, err = s.GetRecipe(context.Background(), rid)
	if !errors.Is(err, ErrRecipeNotFound) {
		t.Fatalf("err = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_GetRecipe_success(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	rid := "550e8400-e29b-41d4-a716-446655440000"
	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	main := sqlmock.NewRows([]string{"id", "name", "description", "category", "image", "created_at", "updated_at"}).
		AddRow(rid, "Cake", "desc", "sweet", "img", ts, ts)
	mock.ExpectQuery("SELECT id::text, name, description, category, image, created_at, updated_at").
		WithArgs(rid).
		WillReturnRows(main)

	ing := sqlmock.NewRows([]string{"quantity", "unit", "name"}).
		AddRow("2", "cup", "sugar")
	mock.ExpectQuery("FROM recipes_ingredients").
		WithArgs(rid).
		WillReturnRows(ing)

	st := sqlmock.NewRows([]string{"instruction"}).
		AddRow("Mix well.")
	mock.ExpectQuery("FROM steps").
		WithArgs(rid).
		WillReturnRows(st)

	got, err := s.GetRecipe(context.Background(), rid)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != rid || got.Name != "Cake" {
		t.Fatalf("got %+v", got)
	}
	if len(got.Ingredients) != 1 || len(got.Instructions) != 1 {
		t.Fatalf("ingredients=%v instructions=%v", got.Ingredients, got.Instructions)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_DeleteRecipe_notFound(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	rid := "550e8400-e29b-41d4-a716-446655440000"
	mock.ExpectExec("DELETE FROM recipes").
		WithArgs(rid).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := s.DeleteRecipe(context.Background(), rid); !errors.Is(err, ErrRecipeNotFound) {
		t.Fatalf("err = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
