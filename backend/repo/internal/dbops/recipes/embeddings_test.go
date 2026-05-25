package recipes

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	types "juancavallotti.com/recipe-types"
	"juancavallotti.com/recipes-repo/internal/embeddings"
)

type stubEmbedder struct {
	vec  []float32
	err  error
	last []string
}

func (s *stubEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	s.last = append(s.last, text)
	return s.vec, s.err
}

func TestRecipeChunks(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		rec  types.Recipe
		want []string
	}{
		{
			name: "full recipe produces three chunks",
			rec: types.Recipe{
				Name:         "Carbonara",
				Description:  "Classic Roman pasta.",
				Ingredients:  []string{"200g spaghetti", "2 eggs", "guanciale"},
				Instructions: []string{"Boil pasta.", "Mix eggs and cheese.", "Toss."},
			},
			want: []string{
				"Carbonara\nClassic Roman pasta.",
				"200g spaghetti\n2 eggs\nguanciale",
				"Boil pasta.\nMix eggs and cheese.\nToss.",
			},
		},
		{
			name: "missing description still yields summary",
			rec:  types.Recipe{Name: "Toast", Ingredients: []string{"bread"}},
			want: []string{"Toast", "bread"},
		},
		{
			name: "empty recipe yields no chunks",
			rec:  types.Recipe{},
			want: []string{},
		},
		{
			name: "whitespace-only fields are dropped",
			rec:  types.Recipe{Name: "  ", Description: "", Ingredients: []string{"  ", "salt"}},
			want: []string{"salt"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := recipeChunks(tc.rec)
			if len(got) != len(tc.want) {
				t.Fatalf("len = %d, want %d (got=%v)", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("chunk %d = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// blockingEmbedder lets a test pause an Embed call until a channel
// fires. Used to assert that Store.Wait blocks until in-flight
// async work has completed.
type blockingEmbedder struct {
	release chan struct{}
	started chan struct{}
	vec     []float32
}

func (b *blockingEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	select {
	case b.started <- struct{}{}:
	default:
	}
	<-b.release
	return b.vec, nil
}

func TestStoreWaitDrainsAsyncIndexing(t *testing.T) {
	t.Parallel()

	vec := make([]float32, embeddings.Dimensions)
	emb := &blockingEmbedder{
		release: make(chan struct{}),
		started: make(chan struct{}, 1),
		vec:     vec,
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db, WithEmbedClient(emb))

	id := "550e8400-e29b-41d4-a716-446655440000"
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	// IndexRecipe's read path
	mock.ExpectQuery(`FROM recipes WHERE id`).WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "image", "created_at", "updated_at"}).
			AddRow(id, "Toast", "", "", "", now, now))
	mock.ExpectQuery("FROM recipes_ingredients").WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"quantity", "unit", "name"}))
	mock.ExpectQuery("FROM steps").WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"instruction"}))
	mock.ExpectQuery("FROM recipes_images").WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "image_base64", "is_featured", "created_at"}))
	// Write path
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM recipe_embeddings").WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO recipe_embeddings").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	s.indexRecipeAsync(context.Background(), id)

	// Goroutine should be running but blocked on the embedder.
	<-emb.started

	waitDone := make(chan struct{})
	go func() {
		s.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		t.Fatal("Wait returned before the embedding goroutine completed")
	case <-time.After(50 * time.Millisecond):
		// Expected — Wait is blocking.
	}

	close(emb.release)

	select {
	case <-waitDone:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("Wait did not return after the embedding goroutine completed")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStoreWaitReturnsImmediatelyWhenNoop(t *testing.T) {
	t.Parallel()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db) // default Noop

	// indexRecipeAsync should short-circuit and not increment the WG.
	s.indexRecipeAsync(context.Background(), "any")

	done := make(chan struct{})
	go func() {
		s.Wait()
		close(done)
	}()
	select {
	case <-done:
		// good
	case <-time.After(time.Second):
		t.Fatal("Wait blocked with no in-flight work and a Noop client")
	}
}

func TestIndexRecipeReturnsErrDisabledWithNoopClient(t *testing.T) {
	t.Parallel()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db) // default Noop client
	if err := s.IndexRecipe(context.Background(), "any"); !errors.Is(err, embeddings.ErrDisabled) {
		t.Fatalf("err = %v, want ErrDisabled", err)
	}
}

func TestIndexRecipePropagatesEmbedderError(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	boom := errors.New("embed boom")
	stub := &stubEmbedder{err: boom}
	s := NewStore(db, WithEmbedClient(stub))

	id := "550e8400-e29b-41d4-a716-446655440000"
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`FROM recipes WHERE id`).WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "category", "image", "created_at", "updated_at"}).
			AddRow(id, "Toast", "", "", "", now, now))
	mock.ExpectQuery("FROM recipes_ingredients").WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"quantity", "unit", "name"}))
	mock.ExpectQuery("FROM steps").WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"instruction"}))
	mock.ExpectQuery("FROM recipes_images").WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "image_base64", "is_featured", "created_at"}))

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM recipe_embeddings").WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 0))
	// Embedder fails on the first chunk; transaction rolls back.
	mock.ExpectRollback()

	if err := s.IndexRecipe(context.Background(), id); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want boom", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
