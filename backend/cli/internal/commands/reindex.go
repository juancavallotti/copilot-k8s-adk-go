package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	repo "juancavallotti.com/recipes-repo"
)

// reindexExitFailed is returned when at least one row in the reindex
// pass failed to embed. Distinct from ErrUsage (exit code 2) so the
// agent can tell "I gave bad arguments" from "some rows failed."
var reindexExitFailed = errors.New("reindex: one or more rows failed")

// cmdReindex implements `recipes-cli reindex --target {recipes|events|all}`.
// Today only the recipes target is wired; Commit 4 will add events and
// all. Output is human-readable by default, JSON Lines with --json so
// the agent can parse progress.
func (r Runner) cmdReindex(ctx context.Context, cmdRepo EmbedRepo, args []string) error {
	const usage = "usage: recipes-cli reindex --target {recipes|events|all} [--force] [--limit N] [--json]"

	var target string
	var force, returnJSON bool
	limit := 0
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--target":
			if i+1 >= len(args) {
				return r.usageError(usage)
			}
			target = args[i+1]
			i++
		case "--force":
			force = true
		case "--json":
			returnJSON = true
		case "--limit":
			if i+1 >= len(args) {
				return r.usageError(usage)
			}
			n, err := strconv.Atoi(args[i+1])
			if err != nil || n < 0 {
				return r.usageError(usage)
			}
			limit = n
			i++
		default:
			return r.usageError(usage)
		}
	}
	if target == "" {
		return r.usageError(usage)
	}

	switch target {
	case "recipes":
		return r.reindexRecipes(ctx, cmdRepo, force, limit, returnJSON)
	case "events":
		return r.reindexEvents(ctx, cmdRepo, force, limit, returnJSON)
	case "all":
		// Run recipes, then events. Aggregate exit-code semantics:
		// if either reports a failed row, return reindexExitFailed.
		errR := r.reindexRecipes(ctx, cmdRepo, force, limit, returnJSON)
		errE := r.reindexEvents(ctx, cmdRepo, force, limit, returnJSON)
		switch {
		case errR != nil && !errors.Is(errR, reindexExitFailed):
			return errR
		case errE != nil && !errors.Is(errE, reindexExitFailed):
			return errE
		case errors.Is(errR, reindexExitFailed) || errors.Is(errE, reindexExitFailed):
			return reindexExitFailed
		}
		return nil
	default:
		return r.usageError(usage)
	}
}

func (r Runner) reindexRecipes(ctx context.Context, cmdRepo EmbedRepo, force bool, limit int, returnJSON bool) error {
	var ok, failed int
	enc := json.NewEncoder(r.stdout)
	onReport := func(rep repo.IndexRecipeReport) {
		switch rep.Status {
		case "ok":
			ok++
		case "error":
			failed++
		}
		if returnJSON {
			_ = enc.Encode(rep)
			return
		}
		if rep.Status == "ok" {
			fmt.Fprintf(r.stdout, "%s\tok\n", rep.ID)
		} else {
			fmt.Fprintf(r.stdout, "%s\terror: %s\n", rep.ID, rep.Error)
		}
	}
	err := cmdRepo.ReindexRecipes(ctx, repo.ReindexOptions{
		Force:    force,
		Limit:    limit,
		OnReport: onReport,
	})
	if err != nil {
		return fmt.Errorf("reindex: %w", err)
	}
	if !returnJSON {
		fmt.Fprintf(r.stdout, "indexed %d recipe(s) (%d ok, %d failed)\n", ok+failed, ok, failed)
	}
	if failed > 0 {
		return reindexExitFailed
	}
	return nil
}

func (r Runner) reindexEvents(ctx context.Context, cmdRepo EmbedRepo, force bool, limit int, returnJSON bool) error {
	var ok, failed int
	enc := json.NewEncoder(r.stdout)
	onReport := func(rep repo.IndexEventReport) {
		switch rep.Status {
		case "ok":
			ok++
		case "error":
			failed++
		}
		if returnJSON {
			_ = enc.Encode(rep)
			return
		}
		if rep.Status == "ok" {
			fmt.Fprintf(r.stdout, "%s\tok\n", rep.ID)
		} else {
			fmt.Fprintf(r.stdout, "%s\terror: %s\n", rep.ID, rep.Error)
		}
	}
	err := cmdRepo.ReindexEvents(ctx, repo.ReindexEventsOptions{
		Force:    force,
		Limit:    limit,
		OnReport: onReport,
	})
	if err != nil {
		return fmt.Errorf("reindex: %w", err)
	}
	if !returnJSON {
		fmt.Fprintf(r.stdout, "indexed %d event(s) (%d ok, %d failed)\n", ok+failed, ok, failed)
	}
	if failed > 0 {
		return reindexExitFailed
	}
	return nil
}
