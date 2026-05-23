package recipes

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
	if _, err := s.AddRecipePhoto(ctx, "550e8400-e29b-41d4-a716-446655440000", types.Photo{}); !errors.Is(err, errNilDB) {
		t.Fatalf("AddRecipePhoto err = %v", err)
	}
	if err := s.DeleteRecipePhoto(ctx, "550e8400-e29b-41d4-a716-446655440000", "650e8400-e29b-41d4-a716-446655440000"); !errors.Is(err, errNilDB) {
		t.Fatalf("DeleteRecipePhoto err = %v", err)
	}
	if err := s.SetFeaturedRecipePhoto(ctx, "550e8400-e29b-41d4-a716-446655440000", "650e8400-e29b-41d4-a716-446655440000"); !errors.Is(err, errNilDB) {
		t.Fatalf("SetFeaturedRecipePhoto err = %v", err)
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

func TestStore_GetRecipes_includesPhotos(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	rid := "550e8400-e29b-41d4-a716-446655440000"
	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	recipes := sqlmock.NewRows([]string{"id", "name", "description", "category", "image", "created_at", "updated_at"}).
		AddRow(rid, "Cake", "desc", "sweet", "", ts, ts)
	mock.ExpectQuery("SELECT id::text").
		WillReturnRows(recipes)

	photos := sqlmock.NewRows([]string{"recipe_id", "id", "image_base64", "is_featured"}).
		AddRow(rid, "650e8400-e29b-41d4-a716-446655440000", "aW1n", true)
	mock.ExpectQuery("FROM recipes_images").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(photos)

	got, err := s.GetRecipes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d", len(got))
	}
	if len(got[0].Photos) != 1 || !got[0].Photos[0].Featured || got[0].Photos[0].ImageBase64 != "aW1n" {
		t.Fatalf("photos=%#v", got[0].Photos)
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

	photos := sqlmock.NewRows([]string{"id", "image_base64", "is_featured"}).
		AddRow("650e8400-e29b-41d4-a716-446655440000", "aW1n", true)
	mock.ExpectQuery("FROM recipes_images").
		WithArgs(rid).
		WillReturnRows(photos)

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
	if len(got.Photos) != 1 || !got.Photos[0].Featured || got.Photos[0].ImageBase64 != "aW1n" {
		t.Fatalf("photos=%#v", got.Photos)
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

func TestStore_DeleteRecipePhoto_notFound(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	rid := "550e8400-e29b-41d4-a716-446655440000"
	pid := "650e8400-e29b-41d4-a716-446655440000"
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM recipes_images").
		WithArgs(rid, pid).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	if err := s.DeleteRecipePhoto(context.Background(), rid, pid); !errors.Is(err, ErrPhotoNotFound) {
		t.Fatalf("err = %v, want ErrPhotoNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_DeleteRecipePhoto_success(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	rid := "550e8400-e29b-41d4-a716-446655440000"
	pid := "650e8400-e29b-41d4-a716-446655440000"
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM recipes_images").
		WithArgs(rid, pid).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM recipe_images").
		WithArgs(pid).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE recipes").
		WithArgs(rid).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := s.DeleteRecipePhoto(context.Background(), rid, pid); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_SetFeaturedRecipePhoto_notFound(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	rid := "550e8400-e29b-41d4-a716-446655440000"
	pid := "650e8400-e29b-41d4-a716-446655440000"
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE recipes_images").
		WithArgs(rid).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE recipes_images").
		WithArgs(rid, pid).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	if err := s.SetFeaturedRecipePhoto(context.Background(), rid, pid); !errors.Is(err, ErrPhotoNotFound) {
		t.Fatalf("err = %v, want ErrPhotoNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_SetFeaturedRecipePhoto_success(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	rid := "550e8400-e29b-41d4-a716-446655440000"
	pid := "650e8400-e29b-41d4-a716-446655440000"
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE recipes_images").
		WithArgs(rid).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE recipes_images").
		WithArgs(rid, pid).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE recipes").
		WithArgs(rid).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := s.SetFeaturedRecipePhoto(context.Background(), rid, pid); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
