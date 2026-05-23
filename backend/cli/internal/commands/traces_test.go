package commands

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	types "juancavallotti.com/recipe-types"
)

func TestRun_LogTraceInsertsRowFromStdin(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	line := `{"time":"2026-05-22T10:00:00Z","level":"INFO","msg":"agent.event","invocation_id":"inv-abc","text":"hello"}`
	r, _, stderr := testRunner(line+"\n", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"log-trace"}); err != nil {
		t.Fatalf("Run log-trace: %v", err)
	}
	if repo.logTraceCalls != 1 {
		t.Fatalf("log trace calls = %d, want 1", repo.logTraceCalls)
	}
	got := repo.logTraceEntries[0]
	if got.eventID != "inv-abc" {
		t.Fatalf("eventID = %q, want inv-abc", got.eventID)
	}
	want := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	if !got.occurredAt.Equal(want) {
		t.Fatalf("occurredAt = %v, want %v", got.occurredAt, want)
	}
	if string(got.data) != line {
		t.Fatalf("data = %s, want %s", got.data, line)
	}
	if !strings.Contains(stderr.String(), "inserted=1 skipped=0") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRun_LogTraceHonorsTimeFromTrace(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, _, _ := testRunner(`{"time":"2020-01-01T00:00:00Z","invocation_id":"inv-old"}`+"\n", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"log-trace"}); err != nil {
		t.Fatalf("Run log-trace: %v", err)
	}
	got := repo.logTraceEntries[0].occurredAt
	want := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("occurredAt = %v, want %v (the trace's own time, not now())", got, want)
	}
}

func TestRun_LogTraceSkipsLinesMissingFields(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	stdin := strings.Join([]string{
		`{"msg":"agent.starting","time":"2026-05-22T10:00:00Z"}`,           // no invocation_id
		`{"time":"2026-05-22T10:00:01Z","msg":"agent.event","invocation_id":"inv-xyz"}`, // good
		`{"msg":"agent.event","invocation_id":"inv-notime"}`,                            // no time
		`{"time":"bogus","invocation_id":"inv-badtime"}`,                                // unparseable time
		`not even json`,                                                                 // bad json
		``,                                                                              // blank
	}, "\n") + "\n"
	r, _, stderr := testRunner(stdin, repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"log-trace"}); err != nil {
		t.Fatalf("Run log-trace: %v", err)
	}
	if repo.logTraceCalls != 1 {
		t.Fatalf("log trace calls = %d, want 1", repo.logTraceCalls)
	}
	if repo.logTraceEntries[0].eventID != "inv-xyz" {
		t.Fatalf("eventID = %q", repo.logTraceEntries[0].eventID)
	}
	if !strings.Contains(stderr.String(), "inserted=1 skipped=4") {
		t.Fatalf("stderr = %q, want inserted=1 skipped=4", stderr.String())
	}
}

func TestRun_LogTraceInsertErrorAbortsLoop(t *testing.T) {
	repo := &fakeRepo{logTraceErr: errors.New("boom")}
	var factoryCalls int
	stdin := `{"time":"2026-05-22T10:00:00Z","invocation_id":"a"}` + "\n" +
		`{"time":"2026-05-22T10:00:01Z","invocation_id":"b"}` + "\n"
	r, _, _ := testRunner(stdin, repo, &factoryCalls)

	err := r.Run(context.Background(), []string{"log-trace"})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v, want wrapped boom", err)
	}
	if repo.logTraceCalls != 1 {
		t.Fatalf("log trace calls = %d, want 1 (loop should abort on first error)", repo.logTraceCalls)
	}
}

func TestRun_LogTraceCustomFieldNames(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	line := `{"ts":"2026-05-22T10:00:00Z","run_id":"run-1","msg":"x"}`
	r, _, _ := testRunner(line+"\n", repo, &factoryCalls)

	err := r.Run(context.Background(), []string{
		"log-trace",
		"--event-id-field", "run_id",
		"--time-field", "ts",
	})
	if err != nil {
		t.Fatalf("Run log-trace: %v", err)
	}
	if repo.logTraceCalls != 1 || repo.logTraceEntries[0].eventID != "run-1" {
		t.Fatalf("entries = %#v", repo.logTraceEntries)
	}
}

func TestRun_LogTraceRejectsUnknownFlag(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, _, stderr := testRunner("", repo, &factoryCalls)

	err := r.Run(context.Background(), []string{"log-trace", "--bogus"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "--event-id-field") {
		t.Fatalf("stderr = %q, want usage mentioning --event-id-field", stderr.String())
	}
	if repo.logTraceCalls != 0 {
		t.Fatalf("log trace calls = %d, want 0", repo.logTraceCalls)
	}
}

func TestRun_LogTraceFlagWithoutValue(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, _, _ := testRunner("", repo, &factoryCalls)

	err := r.Run(context.Background(), []string{"log-trace", "--event-id-field"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
}

func TestRun_ListEventsPrintsJSONLinesAndForwardsDefaults(t *testing.T) {
	ts1 := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 5, 22, 10, 0, 10, 0, time.UTC)
	repo := &fakeRepo{
		listEventsResult: []types.Event{
			{EventID: "inv-a", StartedAt: ts1, EndedAt: ts2, TraceCount: 3},
			{EventID: "inv-b", StartedAt: ts1, EndedAt: ts1, TraceCount: 1},
		},
	}
	var factoryCalls int
	r, stdout, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"list-events"}); err != nil {
		t.Fatalf("Run list-events: %v", err)
	}
	if repo.listEventsCalls != 1 {
		t.Fatalf("list-events calls = %d, want 1", repo.listEventsCalls)
	}
	if repo.listEventsLimit != 50 || repo.listEventsOffset != 0 {
		t.Fatalf("paging = (limit=%d, offset=%d), want (50, 0)", repo.listEventsLimit, repo.listEventsOffset)
	}

	lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %q", len(lines), stdout.String())
	}
	var first types.Event
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("line 0 not JSON: %v\n%s", err, lines[0])
	}
	if first.EventID != "inv-a" || first.TraceCount != 3 {
		t.Fatalf("line 0 = %#v", first)
	}
}

func TestRun_ListEventsForwardsPagingFlags(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, _, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"list-events", "--limit", "25", "--offset", "10"}); err != nil {
		t.Fatalf("Run list-events: %v", err)
	}
	if repo.listEventsLimit != 25 || repo.listEventsOffset != 10 {
		t.Fatalf("paging = (limit=%d, offset=%d), want (25, 10)", repo.listEventsLimit, repo.listEventsOffset)
	}
}

func TestRun_ListEventsRejectsUnknownFlag(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, _, stderr := testRunner("", repo, &factoryCalls)

	err := r.Run(context.Background(), []string{"list-events", "--bogus"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "--limit") {
		t.Fatalf("stderr = %q, want usage mentioning --limit", stderr.String())
	}
	if repo.listEventsCalls != 0 {
		t.Fatalf("list-events calls = %d, want 0", repo.listEventsCalls)
	}
}

func TestRun_ListEventsRejectsNegativeLimit(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, _, _ := testRunner("", repo, &factoryCalls)

	err := r.Run(context.Background(), []string{"list-events", "--limit", "-3"})
	if err == nil || !strings.Contains(err.Error(), "--limit") {
		t.Fatalf("err = %v, want --limit validation error", err)
	}
}

func TestRun_ListTracesPrintsJSONLinesForEvent(t *testing.T) {
	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	repo := &fakeRepo{
		listTracesResult: []types.Trace{
			{ID: "t1", EventID: "inv-a", OccurredAt: ts, Data: json.RawMessage(`{"msg":"agent.event"}`)},
		},
	}
	var factoryCalls int
	r, stdout, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"list-traces", " inv-a "}); err != nil {
		t.Fatalf("Run list-traces: %v", err)
	}
	if repo.listTracesCalls != 1 || repo.listTracesEventID != "inv-a" {
		t.Fatalf("listTraces eventID = %q, calls = %d", repo.listTracesEventID, repo.listTracesCalls)
	}
	if repo.listTracesLimit != 50 || repo.listTracesOffset != 0 {
		t.Fatalf("paging = (limit=%d, offset=%d), want (50, 0)", repo.listTracesLimit, repo.listTracesOffset)
	}

	lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1: %q", len(lines), stdout.String())
	}
	var got types.Trace
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("line 0 not JSON: %v\n%s", err, lines[0])
	}
	if got.ID != "t1" || got.EventID != "inv-a" {
		t.Fatalf("trace = %#v", got)
	}
}

func TestRun_ListTracesForwardsPagingFlags(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, _, _ := testRunner("", repo, &factoryCalls)

	if err := r.Run(context.Background(), []string{"list-traces", "inv-a", "--limit", "10", "--offset", "20"}); err != nil {
		t.Fatalf("Run list-traces: %v", err)
	}
	if repo.listTracesLimit != 10 || repo.listTracesOffset != 20 {
		t.Fatalf("paging = (limit=%d, offset=%d), want (10, 20)", repo.listTracesLimit, repo.listTracesOffset)
	}
}

func TestRun_ListTracesRequiresEventID(t *testing.T) {
	repo := &fakeRepo{}
	var factoryCalls int
	r, _, stderr := testRunner("", repo, &factoryCalls)

	err := r.Run(context.Background(), []string{"list-traces"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "<event-id>") {
		t.Fatalf("stderr = %q, want usage mentioning <event-id>", stderr.String())
	}
	if repo.listTracesCalls != 0 {
		t.Fatalf("listTraces calls = %d, want 0", repo.listTracesCalls)
	}
}
