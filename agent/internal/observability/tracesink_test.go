package observability

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type recordingWriter struct {
	mu    sync.Mutex
	buf   bytes.Buffer
	errOn int // return error on the Nth call (1-based); 0 = never
	calls int
}

func (r *recordingWriter) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	if r.errOn != 0 && r.calls == r.errOn {
		return 0, errors.New("boom")
	}
	return r.buf.Write(p)
}

func (r *recordingWriter) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buf.String()
}

func (r *recordingWriter) Calls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func TestBestEffortWriter_passesThroughOnSuccess(t *testing.T) {
	t.Parallel()
	rw := &recordingWriter{}
	b := &bestEffortWriter{w: rw}

	n, err := b.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != 5 {
		t.Fatalf("n = %d, want 5", n)
	}
	if rw.String() != "hello" {
		t.Fatalf("buf = %q", rw.String())
	}
}

func TestBestEffortWriter_swallowsErrorAndDisablesFutureWrites(t *testing.T) {
	t.Parallel()
	rw := &recordingWriter{errOn: 1}
	var onErrorCalls int
	var onErrorMu sync.Mutex
	b := &bestEffortWriter{w: rw, onError: func(error) {
		onErrorMu.Lock()
		onErrorCalls++
		onErrorMu.Unlock()
	}}

	n, err := b.Write([]byte("first"))
	if err != nil {
		t.Fatalf("Write 1 returned error: %v", err)
	}
	if n != 5 {
		t.Fatalf("n = %d, want 5 (best-effort writers always claim full write)", n)
	}

	n, err = b.Write([]byte("second"))
	if err != nil {
		t.Fatalf("Write 2 returned error: %v", err)
	}
	if n != 6 {
		t.Fatalf("n = %d, want 6", n)
	}
	if rw.Calls() != 1 {
		t.Fatalf("downstream Write called %d times, want 1 (must be disabled after first error)", rw.Calls())
	}
	onErrorMu.Lock()
	defer onErrorMu.Unlock()
	if onErrorCalls != 1 {
		t.Fatalf("onError calls = %d, want 1", onErrorCalls)
	}
}

func TestBestEffortWriter_onErrorFiresOnlyOnce(t *testing.T) {
	t.Parallel()
	rw := &recordingWriter{errOn: 1}
	var onErrorCalls atomicInt
	b := &bestEffortWriter{w: rw, onError: func(error) { onErrorCalls.Add(1) }}

	// Many concurrent writers; only one should trigger onError.
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = b.Write([]byte("x"))
		}()
	}
	wg.Wait()
	if got := onErrorCalls.Load(); got != 1 {
		t.Fatalf("onError calls = %d, want 1", got)
	}
}

type atomicInt struct {
	mu sync.Mutex
	v  int
}

func (a *atomicInt) Add(n int) { a.mu.Lock(); a.v += n; a.mu.Unlock() }
func (a *atomicInt) Load() int { a.mu.Lock(); defer a.mu.Unlock(); return a.v }

func TestStartTraceSink_emptyBinary(t *testing.T) {
	t.Parallel()
	if _, err := StartTraceSink(""); err == nil {
		t.Fatal("expected error for empty binary")
	}
}

func TestStartTraceSink_missingBinary(t *testing.T) {
	t.Parallel()
	if _, err := StartTraceSink("/nonexistent/recipes-cli-binary-totally-not-here"); err == nil {
		t.Fatal("expected error for missing binary")
	}
}

// TestStartTraceSinkCmd_pipesAndDrainsThroughSubprocess uses a small shell
// script that copies stdin to a file. Verifies (a) the sink writer feeds the
// subprocess's stdin, (b) Close() flushes and waits.
func TestStartTraceSinkCmd_pipesAndDrainsThroughSubprocess(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	out := filepath.Join(t.TempDir(), "out.txt")
	cmd := exec.Command("sh", "-c", "cat > "+out)
	sink, err := startTraceSinkCmd(cmd)
	if err != nil {
		t.Fatalf("startTraceSinkCmd: %v", err)
	}
	if _, err := io.WriteString(sink.Writer(), `{"msg":"agent.event","invocation_id":"a"}`+"\n"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := io.WriteString(sink.Writer(), `{"msg":"llm.end","invocation_id":"a"}`+"\n"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := sink.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	gotBytes, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read out: %v", err)
	}
	got := string(gotBytes)
	wantLines := []string{
		`{"msg":"agent.event","invocation_id":"a"}`,
		`{"msg":"llm.end","invocation_id":"a"}`,
	}
	for _, w := range wantLines {
		if !strings.Contains(got, w) {
			t.Fatalf("subprocess output %q missing line %q", got, w)
		}
	}
}

// TestStartTraceSinkCmd_writesAfterSubprocessExitAreSwallowed: subprocess
// exits immediately; subsequent writes through the sink return without error.
func TestStartTraceSinkCmd_writesAfterSubprocessExitAreSwallowed(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("true"); err != nil {
		t.Skip("true not available")
	}
	cmd := exec.Command("true")
	sink, err := startTraceSinkCmd(cmd)
	if err != nil {
		t.Fatalf("startTraceSinkCmd: %v", err)
	}
	// Wait for the subprocess to exit before we write — guarantees the pipe
	// is broken by the time we Write.
	_ = sink.cmd.Wait()

	// First write may or may not fail (race with kernel buffering), but
	// must not panic and must not return an error to the caller.
	if _, err := io.WriteString(sink.Writer(), "ignored\n"); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	// Subsequent writes must definitely be silent no-ops.
	if _, err := io.WriteString(sink.Writer(), "also ignored\n"); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
}

