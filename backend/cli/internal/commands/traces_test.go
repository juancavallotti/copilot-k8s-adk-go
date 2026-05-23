package commands

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
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
