package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (r Runner) cmdLogTrace(ctx context.Context, repo RecipeRepo, args []string) error {
	eventField := "invocation_id"
	timeField := "time"
	usage := "usage: recipes-cli log-trace [--event-id-field <name>] [--time-field <name>]"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--event-id-field":
			if i+1 >= len(args) {
				return r.usageError(usage)
			}
			eventField = args[i+1]
			i++
		case "--time-field":
			if i+1 >= len(args) {
				return r.usageError(usage)
			}
			timeField = args[i+1]
			i++
		default:
			return r.usageError(usage)
		}
	}

	scanner := bufio.NewScanner(r.stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var inserted, skipped int
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			fmt.Fprintf(r.stderr, "log-trace: skipping invalid JSON line: %v\n", err)
			skipped++
			continue
		}

		eventID, ok := readStringField(raw, eventField)
		if !ok {
			fmt.Fprintf(r.stderr, "log-trace: skipping line without/empty %q\n", eventField)
			skipped++
			continue
		}
		tsStr, ok := readStringField(raw, timeField)
		if !ok {
			fmt.Fprintf(r.stderr, "log-trace: skipping line without/empty %q\n", timeField)
			skipped++
			continue
		}
		occurredAt, err := time.Parse(time.RFC3339Nano, tsStr)
		if err != nil {
			fmt.Fprintf(r.stderr, "log-trace: skipping line with unparseable %s=%q: %v\n", timeField, tsStr, err)
			skipped++
			continue
		}

		if err := repo.LogTrace(ctx, eventID, occurredAt, json.RawMessage(line)); err != nil {
			return fmt.Errorf("log-trace: insert: %w", err)
		}
		inserted++
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("log-trace: read stdin: %w", err)
	}
	fmt.Fprintf(r.stderr, "log-trace: inserted=%d skipped=%d\n", inserted, skipped)
	return nil
}

func (r Runner) cmdListEvents(ctx context.Context, repo RecipeRepo, args []string) error {
	const usage = "usage: recipes-cli list-events [--limit N] [--offset N]"
	limit, offset, err := parsePagingFlags(args, usage)
	if err != nil {
		if err == errBadPagingUsage {
			return r.usageError(usage)
		}
		return err
	}
	events, err := repo.ListEvents(ctx, limit, offset)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(r.stdout)
	for _, e := range events {
		if err := enc.Encode(e); err != nil {
			return err
		}
	}
	return nil
}

func (r Runner) cmdListTraces(ctx context.Context, repo RecipeRepo, args []string) error {
	const usage = "usage: recipes-cli list-traces <event-id> [--limit N] [--offset N]"
	if len(args) < 1 {
		return r.usageError(usage)
	}
	eventID := strings.TrimSpace(args[0])
	if eventID == "" {
		return r.usageError(usage)
	}
	limit, offset, err := parsePagingFlags(args[1:], usage)
	if err != nil {
		if err == errBadPagingUsage {
			return r.usageError(usage)
		}
		return err
	}
	traces, err := repo.ListTracesByEvent(ctx, eventID, limit, offset)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(r.stdout)
	for _, t := range traces {
		if err := enc.Encode(t); err != nil {
			return err
		}
	}
	return nil
}

var errBadPagingUsage = fmt.Errorf("bad paging usage")

func parsePagingFlags(args []string, usage string) (limit, offset int, err error) {
	limit = 50
	offset = 0
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--limit":
			if i+1 >= len(args) {
				return 0, 0, errBadPagingUsage
			}
			n, perr := strconv.Atoi(args[i+1])
			if perr != nil || n <= 0 {
				return 0, 0, fmt.Errorf("%s: --limit must be a positive integer", usage)
			}
			limit = n
			i++
		case "--offset":
			if i+1 >= len(args) {
				return 0, 0, errBadPagingUsage
			}
			n, perr := strconv.Atoi(args[i+1])
			if perr != nil || n < 0 {
				return 0, 0, fmt.Errorf("%s: --offset must be a non-negative integer", usage)
			}
			offset = n
			i++
		default:
			return 0, 0, errBadPagingUsage
		}
	}
	return limit, offset, nil
}

func readStringField(raw map[string]json.RawMessage, field string) (string, bool) {
	v, ok := raw[field]
	if !ok {
		return "", false
	}
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return "", false
	}
	s = strings.TrimSpace(s)
	return s, s != ""
}
