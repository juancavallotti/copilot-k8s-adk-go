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

func TestCmdLogTrace_InsertsRowFromStdin(t *testing.T) {
	repo := &fakeRepo{}
	line := `{"time":"2026-05-22T10:00:00Z","level":"INFO","msg":"agent.event","invocation_id":"inv-abc","text":"hello"}`
	r, _, stderr := testRunner(line + "\n")

	if err := r.cmdLogTrace(context.Background(), repo, []string{}); err != nil {
		t.Fatalf("cmdLogTrace: %v", err)
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
	if !strings.Contains(stderr.String(), `"msg":"log-trace.summary"`) ||
		!strings.Contains(stderr.String(), `"inserted":1`) ||
		!strings.Contains(stderr.String(), `"skipped":0`) {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestCmdLogTrace_HonorsTimeFromTrace(t *testing.T) {
	repo := &fakeRepo{}
	r, _, _ := testRunner(`{"time":"2020-01-01T00:00:00Z","invocation_id":"inv-old"}` + "\n")

	if err := r.cmdLogTrace(context.Background(), repo, []string{}); err != nil {
		t.Fatalf("cmdLogTrace: %v", err)
	}
	got := repo.logTraceEntries[0].occurredAt
	want := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("occurredAt = %v, want %v (the trace's own time, not now())", got, want)
	}
}

func TestCmdLogTrace_SkipsLinesMissingFields(t *testing.T) {
	repo := &fakeRepo{}
	stdin := strings.Join([]string{
		`{"msg":"agent.starting","time":"2026-05-22T10:00:00Z"}`,                        // no invocation_id
		`{"time":"2026-05-22T10:00:01Z","msg":"agent.event","invocation_id":"inv-xyz"}`, // good
		`{"msg":"agent.event","invocation_id":"inv-notime"}`,                            // no time
		`{"time":"bogus","invocation_id":"inv-badtime"}`,                                // unparseable time
		`not even json`, // bad json
		``,              // blank
	}, "\n") + "\n"
	r, _, stderr := testRunner(stdin)

	if err := r.cmdLogTrace(context.Background(), repo, []string{}); err != nil {
		t.Fatalf("cmdLogTrace: %v", err)
	}
	if repo.logTraceCalls != 1 {
		t.Fatalf("log trace calls = %d, want 1", repo.logTraceCalls)
	}
	if repo.logTraceEntries[0].eventID != "inv-xyz" {
		t.Fatalf("eventID = %q", repo.logTraceEntries[0].eventID)
	}
	if !strings.Contains(stderr.String(), `"msg":"log-trace.summary"`) ||
		!strings.Contains(stderr.String(), `"inserted":1`) ||
		!strings.Contains(stderr.String(), `"skipped":4`) {
		t.Fatalf("stderr = %q, want structured summary with inserted=1 skipped=4", stderr.String())
	}
}

func TestCmdLogTrace_InsertErrorAbortsLoop(t *testing.T) {
	repo := &fakeRepo{logTraceErr: errors.New("boom")}
	stdin := `{"time":"2026-05-22T10:00:00Z","invocation_id":"a"}` + "\n" +
		`{"time":"2026-05-22T10:00:01Z","invocation_id":"b"}` + "\n"
	r, _, _ := testRunner(stdin)

	err := r.cmdLogTrace(context.Background(), repo, []string{})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v, want wrapped boom", err)
	}
	if repo.logTraceCalls != 1 {
		t.Fatalf("log trace calls = %d, want 1 (loop should abort on first error)", repo.logTraceCalls)
	}
}

func TestCmdLogTrace_CustomFieldNames(t *testing.T) {
	repo := &fakeRepo{}
	line := `{"ts":"2026-05-22T10:00:00Z","run_id":"run-1","msg":"x"}`
	r, _, _ := testRunner(line + "\n")

	err := r.cmdLogTrace(context.Background(), repo, []string{
		"--event-id-field", "run_id",
		"--time-field", "ts",
	})
	if err != nil {
		t.Fatalf("cmdLogTrace: %v", err)
	}
	if repo.logTraceCalls != 1 || repo.logTraceEntries[0].eventID != "run-1" {
		t.Fatalf("entries = %#v", repo.logTraceEntries)
	}
}

func TestCmdLogTrace_RejectsUnknownFlag(t *testing.T) {
	repo := &fakeRepo{}
	r, _, stderr := testRunner("")

	err := r.cmdLogTrace(context.Background(), repo, []string{"--bogus"})
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

func TestCmdLogTrace_FlagWithoutValue(t *testing.T) {
	repo := &fakeRepo{}
	r, _, _ := testRunner("")

	err := r.cmdLogTrace(context.Background(), repo, []string{"--event-id-field"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
}

func TestCmdListEvents_PrintsJSONLinesAndForwardsDefaults(t *testing.T) {
	ts1 := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 5, 22, 10, 0, 10, 0, time.UTC)
	repo := &fakeRepo{
		listEventsResult: []types.Event{
			{EventID: "inv-a", StartedAt: ts1, EndedAt: ts2, TraceCount: 3},
			{EventID: "inv-b", StartedAt: ts1, EndedAt: ts1, TraceCount: 1},
		},
	}
	r, stdout, _ := testRunner("")

	if err := r.cmdListEvents(context.Background(), repo, []string{}); err != nil {
		t.Fatalf("cmdListEvents: %v", err)
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

func TestCmdListEvents_ForwardsPagingFlags(t *testing.T) {
	repo := &fakeRepo{}
	r, _, _ := testRunner("")

	if err := r.cmdListEvents(context.Background(), repo, []string{"--limit", "25", "--offset", "10"}); err != nil {
		t.Fatalf("cmdListEvents: %v", err)
	}
	if repo.listEventsLimit != 25 || repo.listEventsOffset != 10 {
		t.Fatalf("paging = (limit=%d, offset=%d), want (25, 10)", repo.listEventsLimit, repo.listEventsOffset)
	}
}

func TestCmdListEvents_RejectsUnknownFlag(t *testing.T) {
	repo := &fakeRepo{}
	r, _, stderr := testRunner("")

	err := r.cmdListEvents(context.Background(), repo, []string{"--bogus"})
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

func TestCmdListEvents_RejectsNegativeLimit(t *testing.T) {
	repo := &fakeRepo{}
	r, _, _ := testRunner("")

	err := r.cmdListEvents(context.Background(), repo, []string{"--limit", "-3"})
	if err == nil || !strings.Contains(err.Error(), "--limit") {
		t.Fatalf("err = %v, want --limit validation error", err)
	}
}

func TestCmdListTraces_PrintsJSONLinesForEvent(t *testing.T) {
	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	repo := &fakeRepo{
		listTracesResult: []types.Trace{
			{ID: "t1", EventID: "inv-a", OccurredAt: ts, Data: json.RawMessage(`{"msg":"agent.event"}`)},
		},
	}
	r, stdout, _ := testRunner("")

	if err := r.cmdListTraces(context.Background(), repo, []string{" inv-a "}); err != nil {
		t.Fatalf("cmdListTraces: %v", err)
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

func TestCmdListTraces_ForwardsPagingFlags(t *testing.T) {
	repo := &fakeRepo{}
	r, _, _ := testRunner("")

	if err := r.cmdListTraces(context.Background(), repo, []string{"inv-a", "--limit", "10", "--offset", "20"}); err != nil {
		t.Fatalf("cmdListTraces: %v", err)
	}
	if repo.listTracesLimit != 10 || repo.listTracesOffset != 20 {
		t.Fatalf("paging = (limit=%d, offset=%d), want (10, 20)", repo.listTracesLimit, repo.listTracesOffset)
	}
}

func TestCmdListTraces_RequiresEventID(t *testing.T) {
	repo := &fakeRepo{}
	r, _, stderr := testRunner("")

	err := r.cmdListTraces(context.Background(), repo, []string{})
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
