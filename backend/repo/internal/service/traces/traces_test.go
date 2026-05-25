package traces

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	types "juancavallotti.com/recipe-types"
	traceops "juancavallotti.com/recipes-repo/internal/dbops/traces"
)

type fakeStore struct {
	insertTraceCalls   int
	insertTraceEventID string
	insertTraceTime    time.Time
	insertTraceData    json.RawMessage
	insertTraceErr     error
	listEventsLimit    int
	listEventsOffset   int
	listEventsResult   []types.Event
	listTracesCalls    int
	listTracesEventID  string
	listTracesLimit    int
	listTracesOffset   int
	listTracesResult   []types.Trace
	deleteAllCalls     int
	deleteEventID      string
}

func (f *fakeStore) InsertTrace(ctx context.Context, eventID string, occurredAt time.Time, data json.RawMessage) error {
	f.insertTraceCalls++
	f.insertTraceEventID = eventID
	f.insertTraceTime = occurredAt
	f.insertTraceData = data
	return f.insertTraceErr
}

func (f *fakeStore) ListEvents(ctx context.Context, limit, offset int) ([]types.Event, error) {
	f.listEventsLimit = limit
	f.listEventsOffset = offset
	return f.listEventsResult, nil
}

func (f *fakeStore) ListTracesByEvent(ctx context.Context, eventID string, limit, offset int) ([]types.Trace, error) {
	f.listTracesCalls++
	f.listTracesEventID = eventID
	f.listTracesLimit = limit
	f.listTracesOffset = offset
	return f.listTracesResult, nil
}

func (f *fakeStore) DeleteAllEvents(ctx context.Context) error {
	f.deleteAllCalls++
	return nil
}

func (f *fakeStore) DeleteEventByID(ctx context.Context, eventID string) error {
	f.deleteEventID = eventID
	return nil
}

func (f *fakeStore) IndexEvent(ctx context.Context, eventID string, force bool) error {
	return nil
}

func (f *fakeStore) ReindexEvents(ctx context.Context, opts traceops.ReindexEventsOptions) error {
	return nil
}

func (f *fakeStore) SearchEvents(ctx context.Context, query string, limit int) ([]types.EventMatch, error) {
	return nil, nil
}

func (f *fakeStore) Wait() {}

func TestService_LogTrace_rejectsEmptyEventID(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	err := s.LogTrace(context.Background(), "  ", time.Now(), json.RawMessage(`{}`))
	if !errors.Is(err, ErrEmptyEventID) {
		t.Fatalf("err = %v, want ErrEmptyEventID", err)
	}
	if f.insertTraceCalls != 0 {
		t.Fatalf("insertTrace calls = %d, want 0", f.insertTraceCalls)
	}
}

func TestService_LogTrace_rejectsZeroTime(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	err := s.LogTrace(context.Background(), "inv-1", time.Time{}, json.RawMessage(`{}`))
	if !errors.Is(err, ErrZeroOccurredAt) {
		t.Fatalf("err = %v, want ErrZeroOccurredAt", err)
	}
	if f.insertTraceCalls != 0 {
		t.Fatalf("insertTrace calls = %d, want 0", f.insertTraceCalls)
	}
}

func TestService_LogTrace_rejectsEmptyData(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	err := s.LogTrace(context.Background(), "inv-1", time.Now(), json.RawMessage(``))
	if !errors.Is(err, ErrEmptyTraceData) {
		t.Fatalf("err = %v, want ErrEmptyTraceData", err)
	}
	if f.insertTraceCalls != 0 {
		t.Fatalf("insertTrace calls = %d, want 0", f.insertTraceCalls)
	}
}

func TestService_LogTrace_forwardsToStore(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	data := json.RawMessage(`{"msg":"agent.event"}`)

	if err := s.LogTrace(context.Background(), "inv-1", ts, data); err != nil {
		t.Fatalf("LogTrace: %v", err)
	}
	if f.insertTraceCalls != 1 {
		t.Fatalf("insertTrace calls = %d, want 1", f.insertTraceCalls)
	}
	if f.insertTraceEventID != "inv-1" || !f.insertTraceTime.Equal(ts) || string(f.insertTraceData) != string(data) {
		t.Fatalf("forwarded args = %q %v %s", f.insertTraceEventID, f.insertTraceTime, f.insertTraceData)
	}
}

func TestService_LogTrace_propagatesStoreError(t *testing.T) {
	t.Parallel()
	boom := errors.New("boom")
	f := &fakeStore{insertTraceErr: boom}
	s := &Service{store: f}

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
			f := &fakeStore{}
			s := &Service{store: f}
			if _, err := s.ListEvents(context.Background(), tc.limit, tc.offset); err != nil {
				t.Fatal(err)
			}
			if f.listEventsLimit != tc.wantLim || f.listEventsOffset != tc.wantOff {
				t.Fatalf("got (limit=%d, offset=%d), want (%d, %d)", f.listEventsLimit, f.listEventsOffset, tc.wantLim, tc.wantOff)
			}
		})
	}
}

func TestService_ListEvents_returnsStoreResult(t *testing.T) {
	t.Parallel()
	want := []types.Event{{EventID: "inv-a", TraceCount: 3}}
	f := &fakeStore{listEventsResult: want}
	s := &Service{store: f}

	got, err := s.ListEvents(context.Background(), 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].EventID != "inv-a" {
		t.Fatalf("got = %#v", got)
	}
}

func TestService_ListTracesByEvent_rejectsEmptyEventID(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}
	_, err := s.ListTracesByEvent(context.Background(), "", 10, 0)
	if !errors.Is(err, ErrEmptyEventID) {
		t.Fatalf("err = %v, want ErrEmptyEventID", err)
	}
	if f.listTracesCalls != 0 {
		t.Fatalf("listTraces calls = %d, want 0", f.listTracesCalls)
	}
}

func TestService_ListTracesByEvent_forwardsToStore(t *testing.T) {
	t.Parallel()
	want := []types.Trace{{ID: "t1", EventID: "inv-a"}}
	f := &fakeStore{listTracesResult: want}
	s := &Service{store: f}

	got, err := s.ListTracesByEvent(context.Background(), "inv-a", 25, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "t1" {
		t.Fatalf("got = %#v", got)
	}
	if f.listTracesEventID != "inv-a" {
		t.Fatalf("forwarded eventID = %q", f.listTracesEventID)
	}
	if f.listTracesLimit != 25 || f.listTracesOffset != 5 {
		t.Fatalf("forwarded paging = (limit=%d, offset=%d), want (25, 5)", f.listTracesLimit, f.listTracesOffset)
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
			f := &fakeStore{}
			s := &Service{store: f}
			if _, err := s.ListTracesByEvent(context.Background(), "inv-a", tc.limit, tc.offset); err != nil {
				t.Fatal(err)
			}
			if f.listTracesLimit != tc.wantLim || f.listTracesOffset != tc.wantOff {
				t.Fatalf("got (limit=%d, offset=%d), want (%d, %d)", f.listTracesLimit, f.listTracesOffset, tc.wantLim, tc.wantOff)
			}
		})
	}
}

func TestService_DeleteAllEvents_DelegatesToStore(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}

	if err := s.DeleteAllEvents(context.Background()); err != nil {
		t.Fatal(err)
	}
	if f.deleteAllCalls != 1 {
		t.Fatalf("deleteAll calls = %d, want 1", f.deleteAllCalls)
	}
}

func TestService_DeleteEvent_rejectsEmptyEventID(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}

	err := s.DeleteEvent(context.Background(), " ")
	if !errors.Is(err, ErrEmptyEventID) {
		t.Fatalf("err = %v, want ErrEmptyEventID", err)
	}
	if f.deleteEventID != "" {
		t.Fatalf("store should not be called, eventID = %q", f.deleteEventID)
	}
}

func TestService_DeleteEvent_DelegatesToStore(t *testing.T) {
	t.Parallel()
	f := &fakeStore{}
	s := &Service{store: f}

	if err := s.DeleteEvent(context.Background(), "event-1"); err != nil {
		t.Fatal(err)
	}
	if f.deleteEventID != "event-1" {
		t.Fatalf("delete eventID = %q, want event-1", f.deleteEventID)
	}
}
