package commands

import (
	"context"
	"errors"
	"strings"
	"testing"

	types "juancavallotti.com/recipe-types"
	repo "juancavallotti.com/recipes-repo"
)

func TestRun_ListSkillsPrintsJSONL(t *testing.T) {
	r := &fakeRepo{listSkillsResult: []types.Skill{
		{Name: "recipe-management", Description: "Manage recipes.", Content: "ignored on list"},
		{Name: "trace-analysis", Description: "Analyze traces.", Content: "ignored on list"},
	}}
	var factoryCalls int
	runner, stdout, _ := testRunner("", r, &factoryCalls)

	if err := runner.Run(context.Background(), []string{"list-skills"}); err != nil {
		t.Fatalf("Run list-skills: %v", err)
	}
	if r.listSkillsCalls != 1 {
		t.Fatalf("listSkillsCalls = %d, want 1", r.listSkillsCalls)
	}
	lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2; output=%q", len(lines), stdout.String())
	}
	if !strings.Contains(lines[0], `"name":"recipe-management"`) ||
		!strings.Contains(lines[0], `"description":"Manage recipes."`) {
		t.Fatalf("line 0 = %q", lines[0])
	}
	if strings.Contains(lines[0], "content") {
		t.Fatalf("list-skills must omit content; line=%q", lines[0])
	}
}

func TestRun_ListSkillsRejectsExtraArg(t *testing.T) {
	r := &fakeRepo{}
	var factoryCalls int
	runner, _, stderr := testRunner("", r, &factoryCalls)

	err := runner.Run(context.Background(), []string{"list-skills", "extra"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "usage: recipes-cli list-skills") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRun_LoadSkillPrintsContent(t *testing.T) {
	want := "# Recipe Management\n\nDo the thing.\n"
	r := &fakeRepo{getSkillByNameResult: types.Skill{
		Name: "recipe-management", Description: "Manage recipes.", Content: want,
	}}
	var factoryCalls int
	runner, stdout, _ := testRunner("", r, &factoryCalls)

	if err := runner.Run(context.Background(), []string{"load-skill", "recipe-management"}); err != nil {
		t.Fatalf("Run load-skill: %v", err)
	}
	if r.getSkillByNameArg != "recipe-management" {
		t.Fatalf("arg = %q", r.getSkillByNameArg)
	}
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRun_LoadSkillMissingArg(t *testing.T) {
	r := &fakeRepo{}
	var factoryCalls int
	runner, _, stderr := testRunner("", r, &factoryCalls)

	err := runner.Run(context.Background(), []string{"load-skill"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "usage: recipes-cli load-skill <name>") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRun_LoadSkillNotFound(t *testing.T) {
	r := &fakeRepo{getSkillByNameErr: repo.ErrSkillNotFound}
	var factoryCalls int
	runner, stdout, stderr := testRunner("", r, &factoryCalls)

	err := runner.Run(context.Background(), []string{"load-skill", "bogus"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty on not-found", stdout.String())
	}
	if !strings.Contains(stderr.String(), `no skill named "bogus"`) {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
