package observability

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync/atomic"
)

// TraceSink owns a long-running `<binary> log-trace` subprocess that the
// agent feeds slog JSON records into. Each line the agent writes becomes one
// row in the traces table (and an upsert into events).
//
// Failures of the subprocess (binary missing, DB down, broken pipe) never
// propagate back into the agent's hot path: writes are best-effort and silent
// once the pipe goes bad — stdout logging continues unaffected.
type TraceSink struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	writer io.Writer
}

// StartTraceSink spawns `<binary> log-trace`, wires its stdin to the returned
// writer, and forwards its stdout/stderr to ours (so per-batch
// "inserted=N skipped=M" reports still surface). Returns an error if the
// process cannot be started; the caller can then continue without piping.
func StartTraceSink(binary string) (*TraceSink, error) {
	if binary == "" {
		return nil, errors.New("tracesink: empty binary")
	}
	return startTraceSinkCmd(exec.Command(binary, "log-trace"))
}

func startTraceSinkCmd(cmd *exec.Cmd) (*TraceSink, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("tracesink: stdin pipe: %w", err)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("tracesink: start %s: %w", cmd.Path, err)
	}
	s := &TraceSink{cmd: cmd, stdin: stdin}
	s.writer = &bestEffortWriter{
		w: stdin,
		onError: func(err error) {
			log.Printf("tracesink: pipe write error, disabling further writes: %v", err)
		},
	}
	return s, nil
}

// Writer returns the io.Writer that the slog handler should tee into. It is
// safe for concurrent use because slog serializes writes through its own
// mutex. Writes to this writer never return an error.
func (s *TraceSink) Writer() io.Writer { return s.writer }

// Close shuts the subprocess down gracefully: close stdin (signals EOF →
// the CLI flushes its last batch and exits), then wait for the process. Safe
// to call multiple times.
func (s *TraceSink) Close() error {
	if s == nil {
		return nil
	}
	closeErr := s.stdin.Close()
	waitErr := s.cmd.Wait()
	if waitErr != nil {
		return waitErr
	}
	return closeErr
}

// bestEffortWriter swallows write errors so a broken subprocess pipe never
// stalls the agent's logging. The first error disables the writer
// permanently; subsequent writes are no-ops. We log via the stdlib log
// package (not slog) so the error notice cannot recurse back through this
// writer.
type bestEffortWriter struct {
	w       io.Writer
	onError func(error)
	failed  atomic.Bool
}

func (b *bestEffortWriter) Write(p []byte) (int, error) {
	if b.failed.Load() {
		return len(p), nil
	}
	n, err := b.w.Write(p)
	if err != nil {
		if b.failed.CompareAndSwap(false, true) && b.onError != nil {
			b.onError(err)
		}
		return len(p), nil
	}
	return n, nil
}
