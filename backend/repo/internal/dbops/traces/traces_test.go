package traces

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestStore_InsertTrace_nilDB(t *testing.T) {
	t.Parallel()
	s := &Store{db: nil}
	err := s.InsertTrace(context.Background(), "inv-1", time.Now(), json.RawMessage(`{}`))
	if !errors.Is(err, errNilDB) {
		t.Fatalf("err = %v, want errNilDB", err)
	}
}

func TestStore_InsertTrace_contextCanceled(t *testing.T) {
	t.Parallel()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = s.InsertTrace(ctx, "inv-1", time.Now(), json.RawMessage(`{}`))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestStore_InsertTrace_success(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	data := json.RawMessage(`{"msg":"agent.event","invocation_id":"inv-abc"}`)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO events").
		WithArgs("inv-abc", ts, "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO traces").
		WithArgs("inv-abc", ts, []byte(data)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := s.InsertTrace(context.Background(), "inv-abc", ts, data); err != nil {
		t.Fatalf("InsertTrace: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_InsertTrace_convertsToUTC(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	nyTime := time.Date(2026, 5, 22, 6, 0, 0, 0, loc)
	wantUTC := nyTime.UTC() // 10:00:00 UTC

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO events").
		WithArgs("inv-tz", wantUTC, "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO traces").
		WithArgs("inv-tz", wantUTC, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := s.InsertTrace(context.Background(), "inv-tz", nyTime, json.RawMessage(`{}`)); err != nil {
		t.Fatalf("InsertTrace: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_InsertTrace_persistsUserPrompt(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	data := json.RawMessage(`{"msg":"llm.start","user_prompt":"make me pasta"}`)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO events").
		WithArgs("inv-p", ts, "make me pasta").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO traces").
		WithArgs("inv-p", ts, []byte(data)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := s.InsertTrace(context.Background(), "inv-p", ts, data); err != nil {
		t.Fatalf("InsertTrace: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_InsertTrace_unwrapsUserMessage(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	data := json.RawMessage(`{"user_prompt":"{\"userMessage\":\"hello world\"}"}`)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO events").
		WithArgs("inv-u", ts, "hello world").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO traces").
		WithArgs("inv-u", ts, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := s.InsertTrace(context.Background(), "inv-u", ts, data); err != nil {
		t.Fatalf("InsertTrace: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_InsertTrace_rollsBackOnTraceInsertError(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	boom := errors.New("boom")

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO events").
		WithArgs("inv-x", ts, "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO traces").
		WithArgs("inv-x", ts, sqlmock.AnyArg()).
		WillReturnError(boom)
	mock.ExpectRollback()

	if err := s.InsertTrace(context.Background(), "inv-x", ts, json.RawMessage(`{}`)); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want boom", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_ListEvents_nilDB(t *testing.T) {
	t.Parallel()
	s := &Store{db: nil}
	if _, err := s.ListEvents(context.Background(), 10, 0); !errors.Is(err, errNilDB) {
		t.Fatalf("err = %v, want errNilDB", err)
	}
}

func TestStore_ListEvents_empty(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectQuery("FROM events").
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"event_id", "started_at", "ended_at", "trace_count", "user_prompt"}))

	out, err := s.ListEvents(context.Background(), 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("len = %d, want 0", len(out))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_ListEvents_returnsRows(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	t1 := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 5, 22, 10, 0, 10, 0, time.UTC)

	mock.ExpectQuery("FROM events").
		WithArgs(50, 0).
		WillReturnRows(sqlmock.NewRows([]string{"event_id", "started_at", "ended_at", "trace_count", "user_prompt"}).
			AddRow("inv-a", t1, t2, 3, "make pasta").
			AddRow("inv-b", t1, t1, 1, ""))

	out, err := s.ListEvents(context.Background(), 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}
	if out[0].EventID != "inv-a" || out[0].TraceCount != 3 {
		t.Fatalf("out[0] = %#v", out[0])
	}
	if !out[0].StartedAt.Equal(t1) || !out[0].EndedAt.Equal(t2) {
		t.Fatalf("out[0] times = %v..%v", out[0].StartedAt, out[0].EndedAt)
	}
	if out[0].UserPrompt != "make pasta" {
		t.Fatalf("out[0].UserPrompt = %q, want %q", out[0].UserPrompt, "make pasta")
	}
	if out[1].UserPrompt != "" {
		t.Fatalf("out[1].UserPrompt = %q, want empty", out[1].UserPrompt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_ListTracesByEvent_nilDB(t *testing.T) {
	t.Parallel()
	s := &Store{db: nil}
	if _, err := s.ListTracesByEvent(context.Background(), "inv-1", 50, 0); !errors.Is(err, errNilDB) {
		t.Fatalf("err = %v, want errNilDB", err)
	}
}

func TestStore_ListTracesByEvent_returnsRows(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	tid := "550e8400-e29b-41d4-a716-446655440000"
	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	payload := []byte(`{"msg":"agent.event"}`)

	mock.ExpectQuery("FROM traces").
		WithArgs("inv-a", 50, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "event_id", "occurred_at", "data"}).
			AddRow(tid, "inv-a", ts, payload))

	out, err := s.ListTracesByEvent(context.Background(), "inv-a", 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("len = %d, want 1", len(out))
	}
	if out[0].ID != tid || out[0].EventID != "inv-a" {
		t.Fatalf("out[0] = %#v", out[0])
	}
	if string(out[0].Data) != string(payload) {
		t.Fatalf("data = %s, want %s", out[0].Data, payload)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_DeleteAllEvents_nilDB(t *testing.T) {
	t.Parallel()
	s := &Store{db: nil}
	if err := s.DeleteAllEvents(context.Background()); !errors.Is(err, errNilDB) {
		t.Fatalf("err = %v, want errNilDB", err)
	}
}

func TestStore_DeleteAllEvents_success(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectExec("DELETE FROM events").
		WillReturnResult(sqlmock.NewResult(0, 3))

	if err := s.DeleteAllEvents(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_DeleteEventByID_nilDB(t *testing.T) {
	t.Parallel()
	s := &Store{db: nil}
	if err := s.DeleteEventByID(context.Background(), "event-1"); !errors.Is(err, errNilDB) {
		t.Fatalf("err = %v, want errNilDB", err)
	}
}

func TestStore_DeleteEventByID_emptyEventID(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	err = s.DeleteEventByID(context.Background(), " ")
	if !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("err = %v, want ErrEventNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_DeleteEventByID_notFound(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectExec("DELETE FROM events").
		WithArgs("event-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = s.DeleteEventByID(context.Background(), "event-1")
	if !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("err = %v, want ErrEventNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_DeleteEventByID_success(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)

	mock.ExpectExec("DELETE FROM events").
		WithArgs("event-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := s.DeleteEventByID(context.Background(), "event-1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
