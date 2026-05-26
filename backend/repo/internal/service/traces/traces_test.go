package traces

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	types "juancavallotti.com/recipe-types"
	traceops "juancavallotti.com/recipes-repo/internal/dbops/traces"
)

func newMockService(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	return NewService(traceops.NewStore(db)), mock, func() { db.Close() }
}

func TestService_LogTrace_rejectsEmptyEventID(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	err := s.LogTrace(context.Background(), "  ", time.Now(), json.RawMessage(`{}`))
	if !errors.Is(err, ErrEmptyEventID) {
		t.Fatalf("err = %v, want ErrEmptyEventID", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_LogTrace_rejectsZeroTime(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	err := s.LogTrace(context.Background(), "inv-1", time.Time{}, json.RawMessage(`{}`))
	if !errors.Is(err, ErrZeroOccurredAt) {
		t.Fatalf("err = %v, want ErrZeroOccurredAt", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_LogTrace_rejectsEmptyData(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	err := s.LogTrace(context.Background(), "inv-1", time.Now(), json.RawMessage(``))
	if !errors.Is(err, ErrEmptyTraceData) {
		t.Fatalf("err = %v, want ErrEmptyTraceData", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_LogTrace_forwardsToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	data := json.RawMessage(`{"msg":"agent.event"}`)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO events").
		WithArgs("inv-1", ts, "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO traces").
		WithArgs("inv-1", ts, []byte(data)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := s.LogTrace(context.Background(), "inv-1", ts, data); err != nil {
		t.Fatalf("LogTrace: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_LogTrace_propagatesStoreError(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	boom := errors.New("boom")
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO events").WillReturnError(boom)
	mock.ExpectRollback()

	err := s.LogTrace(context.Background(), "inv-1", time.Now(), json.RawMessage(`{}`))
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want boom", err)
	}
}

func TestService_ListEvents_defaultsLimitWhenInvalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		limit   int
		offset  int
		wantLim int
		wantOff int
	}{
		{"zero limit defaults to 50", 0, 0, 50, 0},
		{"negative limit defaults to 50", -5, 0, 50, 0},
		{"over-cap limit defaults to 50", 201, 0, 50, 0},
		{"in-range limit kept", 100, 7, 100, 7},
		{"negative offset clamped to 0", 10, -3, 10, 0},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s, mock, cleanup := newMockService(t)
			defer cleanup()

			mock.ExpectQuery("FROM events").
				WithArgs(tc.wantLim, tc.wantOff).
				WillReturnRows(sqlmock.NewRows([]string{"event_id", "started_at", "ended_at", "trace_count", "user_prompt"}))

			if _, err := s.ListEvents(context.Background(), tc.limit, tc.offset); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestService_ListEvents_returnsStoreResult(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery("FROM events").
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"event_id", "started_at", "ended_at", "trace_count", "user_prompt"}).
			AddRow("inv-a", ts, ts, 3, ""))

	got, err := s.ListEvents(context.Background(), 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].EventID != "inv-a" || got[0].TraceCount != 3 {
		t.Fatalf("got = %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_ListTracesByEvent_rejectsEmptyEventID(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	_, err := s.ListTracesByEvent(context.Background(), "", 10, 0)
	if !errors.Is(err, ErrEmptyEventID) {
		t.Fatalf("err = %v, want ErrEmptyEventID", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_ListTracesByEvent_forwardsToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	want := types.Trace{ID: "t1", EventID: "inv-a", OccurredAt: ts, Data: json.RawMessage(`{}`)}
	mock.ExpectQuery("FROM traces").
		WithArgs("inv-a", 25, 5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "event_id", "occurred_at", "data"}).
			AddRow(want.ID, want.EventID, want.OccurredAt, []byte(want.Data)))

	got, err := s.ListTracesByEvent(context.Background(), "inv-a", 25, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "t1" || got[0].EventID != "inv-a" {
		t.Fatalf("got = %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_ListTracesByEvent_defaultsLimitWhenInvalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		limit   int
		offset  int
		wantLim int
		wantOff int
	}{
		{"zero limit defaults to 50", 0, 0, 50, 0},
		{"negative limit defaults to 50", -5, 0, 50, 0},
		{"over-cap limit defaults to 50", 201, 0, 50, 0},
		{"in-range limit kept", 100, 7, 100, 7},
		{"negative offset clamped to 0", 10, -3, 10, 0},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s, mock, cleanup := newMockService(t)
			defer cleanup()

			mock.ExpectQuery("FROM traces").
				WithArgs("inv-a", tc.wantLim, tc.wantOff).
				WillReturnRows(sqlmock.NewRows([]string{"id", "event_id", "occurred_at", "data"}))

			if _, err := s.ListTracesByEvent(context.Background(), "inv-a", tc.limit, tc.offset); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestService_DeleteAllEvents_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	mock.ExpectExec("DELETE FROM events").WillReturnResult(sqlmock.NewResult(0, 0))

	if err := s.DeleteAllEvents(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_DeleteEvent_rejectsEmptyEventID(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	err := s.DeleteEvent(context.Background(), " ")
	if !errors.Is(err, ErrEmptyEventID) {
		t.Fatalf("err = %v, want ErrEmptyEventID", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestService_DeleteEvent_DelegatesToStore(t *testing.T) {
	t.Parallel()
	s, mock, cleanup := newMockService(t)
	defer cleanup()

	mock.ExpectExec("DELETE FROM events").
		WithArgs("event-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := s.DeleteEvent(context.Background(), "event-1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
