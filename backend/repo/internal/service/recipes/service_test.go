package recipes

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	types "juancavallotti.com/recipe-types"

	recipeops "juancavallotti.com/recipes-repo/internal/dbops/recipes"
)

const testUUID = "550e8400-e29b-41d4-a716-446655440000"
const photoUUID = "650e8400-e29b-41d4-a716-446655440000"

func newMockService(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	return NewService(recipeops.NewStore(db)), mock, func() { db.Close() }
}

// expectCreateRecipeSQL sets up sqlmock expectations matching the SQL the
// store emits for CreateRecipe / CreateRecipeWithID on a recipe with a
// single "i" ingredient, single "s" instruction, no photos.
func expectCreateRecipeSQL(mock sqlmock.Sqlmock, returnedID string, explicit bool) {
	mock.ExpectBegin()
	if explicit {
		mock.ExpectExec("INSERT INTO recipes").
			WillReturnResult(sqlmock.NewResult(0, 1))
	} else {
		mock.ExpectQuery("INSERT INTO recipes").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(returnedID))
	}
	mock.ExpectQuery("INSERT INTO ingredients").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec("INSERT INTO recipes_ingredients").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO steps").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM recipes_images").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
}

func expectGetRecipeSQL(mock sqlmock.Sqlmock, id string) {
	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM recipes WHERE id").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "image", "created_at", "updated_at"}).
			AddRow(id, "from-store", "", "", "", ts, ts))
	mock.ExpectQuery("FROM recipes_ingredients").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"quantity", "unit", "name"}))
	mock.ExpectQuery("FROM steps").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"instruction"}))
	mock.ExpectQuery("FROM recipes_images").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "image_base64", "is_featured"}))
}

func expectUpdateRecipeSQL(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE recipes").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM recipes_ingredients").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM steps").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("INSERT INTO ingredients").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec("INSERT INTO recipes_ingredients").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO steps").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM recipes_images").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
}

func TestService_GetRecipes_NoValidation(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	mock.ExpectQuery("FROM recipes").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "image", "created_at", "updated_at"}))

	got, err := s.GetRecipes(context.Background())
	if err != nil {
		t.Fatalf("GetRecipes: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_GetRecipe_ValidationShortCircuit(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	_, err := s.GetRecipe(context.Background(), "  ")
	if !errors.Is(err, ErrInvalidRecipeID) {
		t.Fatalf("err = %v, want ErrInvalidRecipeID", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_GetRecipe_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	expectGetRecipeSQL(mock, testUUID)

	got, err := s.GetRecipe(context.Background(), testUUID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != testUUID || got.Name != "from-store" {
		t.Fatalf("got %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_CreateRecipe_ValidationShortCircuit(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	_, err := s.CreateRecipe(context.Background(), types.Recipe{Name: ""})
	if !errors.Is(err, ErrInvalidRecipe) {
		t.Fatalf("err = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_CreateRecipe_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	expectCreateRecipeSQL(mock, testUUID, false)

	r := types.Recipe{Name: "n", Ingredients: []string{"i"}, Instructions: []string{"s"}}
	id, err := s.CreateRecipe(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	if id != testUUID {
		t.Fatalf("id = %q, want %q", id, testUUID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_UpdateRecipe_ValidationShortCircuit(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	err := s.UpdateRecipe(context.Background(), types.Recipe{ID: ""})
	if !errors.Is(err, ErrInvalidRecipeID) {
		t.Fatalf("err = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_DeleteRecipe_ValidationShortCircuit(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	if err := s.DeleteRecipe(context.Background(), ""); !errors.Is(err, ErrInvalidRecipeID) {
		t.Fatalf("err = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_AddRecipePhoto_ValidatesBase64(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	_, err := s.AddRecipePhoto(context.Background(), testUUID, types.Photo{ImageBase64: "not base64"})
	if !errors.Is(err, ErrInvalidRecipe) {
		t.Fatalf("err = %v, want ErrInvalidRecipe", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_AddRecipePhoto_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(testUUID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("UPDATE recipes_images").
		WithArgs(testUUID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("INSERT INTO recipe_images").
		WithArgs("aW1n").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(photoUUID))
	mock.ExpectExec("INSERT INTO recipes_images").
		WithArgs(testUUID, photoUUID, true).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE recipes").
		WithArgs(testUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	id, err := s.AddRecipePhoto(context.Background(), testUUID, types.Photo{ImageBase64: "aW1n", Featured: true})
	if err != nil {
		t.Fatal(err)
	}
	if id != photoUUID {
		t.Fatalf("id = %q, want %q", id, photoUUID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_DeleteRecipePhoto_ValidationShortCircuit(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	err := s.DeleteRecipePhoto(context.Background(), testUUID, " ")
	if !errors.Is(err, ErrInvalidRecipeID) {
		t.Fatalf("err = %v, want ErrInvalidRecipeID", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_DeleteRecipePhoto_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM recipes_images").
		WithArgs(testUUID, photoUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM recipe_images").
		WithArgs(photoUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE recipes").
		WithArgs(testUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := s.DeleteRecipePhoto(context.Background(), testUUID, photoUUID); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_SetFeaturedRecipePhoto_ValidationShortCircuit(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	err := s.SetFeaturedRecipePhoto(context.Background(), testUUID, " ")
	if !errors.Is(err, ErrInvalidRecipeID) {
		t.Fatalf("err = %v, want ErrInvalidRecipeID", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_SetFeaturedRecipePhoto_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE recipes_images").
		WithArgs(testUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE recipes_images").
		WithArgs(testUUID, photoUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE recipes").
		WithArgs(testUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := s.SetFeaturedRecipePhoto(context.Background(), testUUID, photoUUID); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_CreateRecipe_StoreErrorPropagates(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	want := errors.New("db down")
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO recipes").WillReturnError(want)
	mock.ExpectRollback()

	r := types.Recipe{Name: "n", Ingredients: []string{"i"}, Instructions: []string{"s"}}
	_, err := s.CreateRecipe(context.Background(), r)
	if !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}
}

func TestService_ImportRecipe_EmptyID_Creates(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	expectCreateRecipeSQL(mock, testUUID, false)

	r := types.Recipe{Name: "n", Ingredients: []string{"i"}, Instructions: []string{"s"}}
	if err := s.ImportRecipe(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_ImportRecipe_NewUUID_InsertsWithID(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	// GetRecipe returns ErrRecipeNotFound: only the main SELECT runs and errors.
	mock.ExpectQuery("FROM recipes WHERE id").
		WithArgs(testUUID).
		WillReturnError(sql.ErrNoRows)

	expectCreateRecipeSQL(mock, testUUID, true)

	r := types.Recipe{ID: testUUID, Name: "n", Ingredients: []string{"i"}, Instructions: []string{"s"}}
	if err := s.ImportRecipe(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_ImportRecipe_ExistingUUID_Updates(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	expectGetRecipeSQL(mock, testUUID)
	expectUpdateRecipeSQL(mock)

	r := types.Recipe{ID: testUUID, Name: "n", Ingredients: []string{"i"}, Instructions: []string{"s"}}
	if err := s.ImportRecipe(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
