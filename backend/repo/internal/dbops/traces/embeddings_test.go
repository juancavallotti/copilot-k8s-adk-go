package traces

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"juancavallotti.com/recipes-repo/internal/embeddings"
)

type stubEmbedder struct {
	vec []float32
	err error
}

func (s *stubEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	return s.vec, s.err
}

func makeVec() []float32 {
	v := make([]float32, embeddings.Dimensions)
	for i := range v {
		v[i] = float32(i) / 1000.0
	}
	return v
}

func TestIndexEventWithNoopReturnsErrDisabled(t *testing.T) {
	t.Parallel()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)
	err = s.IndexEvent(context.Background(), "inv-1", false)
	if !errors.Is(err, embeddings.ErrDisabled) {
		t.Fatalf("err = %v, want ErrDisabled", err)
	}
}

func TestIndexEventEmptyPromptIsNoop(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db, WithEmbedClient(&stubEmbedder{vec: makeVec()}))

	mock.ExpectQuery("SELECT user_prompt FROM events").
		WithArgs("inv-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_prompt"}).AddRow(""))

	if err := s.IndexEvent(context.Background(), "inv-1", false); err != nil {
		t.Fatalf("IndexEvent: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestIndexEventSkipsWhenAlreadyIndexed(t *testing.T) {
	t.Parallel()
	emb := &stubEmbedder{vec: makeVec()}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db, WithEmbedClient(emb))

	mock.ExpectQuery("SELECT user_prompt FROM events").
		WithArgs("inv-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_prompt"}).AddRow("make pasta"))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("inv-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	if err := s.IndexEvent(context.Background(), "inv-1", false); err != nil {
		t.Fatalf("IndexEvent: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestIndexEventForceUpserts(t *testing.T) {
	t.Parallel()
	emb := &stubEmbedder{vec: makeVec()}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db, WithEmbedClient(emb))

	mock.ExpectQuery("SELECT user_prompt FROM events").
		WithArgs("inv-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_prompt"}).AddRow("make pasta"))
	mock.ExpectExec("INSERT INTO event_embeddings").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := s.IndexEvent(context.Background(), "inv-1", true); err != nil {
		t.Fatalf("IndexEvent force: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInsertTraceFiresAsyncIndexingWhenPromptPresent(t *testing.T) {
	t.Parallel()
	emb := &stubEmbedder{vec: makeVec()}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db, WithEmbedClient(emb))

	ts := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	data := json.RawMessage(`{"msg":"llm.start","user_prompt":"hello"}`)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO events").
		WithArgs("inv-1", ts, "hello").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO traces").
		WithArgs("inv-1", ts, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Async goroutine: IndexEvent reads user_prompt, checks existence,
	// embeds, upserts.
	mock.ExpectQuery("SELECT user_prompt FROM events").
		WithArgs("inv-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_prompt"}).AddRow("hello"))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("inv-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("INSERT INTO event_embeddings").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := s.InsertTrace(context.Background(), "inv-1", ts, data); err != nil {
		t.Fatalf("InsertTrace: %v", err)
	}
	s.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInsertTraceDoesNotFireWhenNoPrompt(t *testing.T) {
	t.Parallel()
	emb := &stubEmbedder{vec: makeVec()}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db, WithEmbedClient(emb))

	ts := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	data := json.RawMessage(`{"msg":"agent.event"}`) // no user_prompt

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO events").
		WithArgs("inv-1", ts, "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO traces").
		WithArgs("inv-1", ts, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	// No async embedding expectations — the hook must short-circuit.

	if err := s.InsertTrace(context.Background(), "inv-1", ts, data); err != nil {
		t.Fatalf("InsertTrace: %v", err)
	}
	s.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
