package commands

import (
	"context"
	"errors"
	"strings"
	"testing"

	types "juancavallotti.com/recipe-types"
	repo "juancavallotti.com/recipes-repo"
)

func TestCmdListSkills_PrintsJSONL(t *testing.T) {
	fake := &fakeRepo{listSkillsResult: []types.Skill{
		{Name: "recipe-management", Description: "Manage recipes.", Content: "ignored on list"},
		{Name: "trace-analysis", Description: "Analyze traces.", Content: "ignored on list"},
	}}
	runner, stdout, _ := testRunner("")

	if err := runner.cmdListSkills(context.Background(), fake); err != nil {
		t.Fatalf("cmdListSkills: %v", err)
	}
	if fake.listSkillsCalls != 1 {
		t.Fatalf("listSkillsCalls = %d, want 1", fake.listSkillsCalls)
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
	runner, _, stderr := testRunner("")

	err := runner.Run(context.Background(), []string{"list-skills", "extra"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "usage: recipes-cli list-skills") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestCmdLoadSkill_PrintsContent(t *testing.T) {
	want := "# Recipe Management\n\nDo the thing.\n"
	fake := &fakeRepo{getSkillByNameResult: types.Skill{
		Name: "recipe-management", Description: "Manage recipes.", Content: want,
	}}
	runner, stdout, _ := testRunner("")

	if err := runner.cmdLoadSkill(context.Background(), fake, "recipe-management"); err != nil {
		t.Fatalf("cmdLoadSkill: %v", err)
	}
	if fake.getSkillByNameArg != "recipe-management" {
		t.Fatalf("arg = %q", fake.getSkillByNameArg)
	}
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRun_LoadSkillMissingArg(t *testing.T) {
	runner, _, stderr := testRunner("")

	err := runner.Run(context.Background(), []string{"load-skill"})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("err = %v, want ErrUsage", err)
	}
	if !strings.Contains(stderr.String(), "usage: recipes-cli load-skill <name>") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestCmdLoadSkill_NotFound(t *testing.T) {
	fake := &fakeRepo{getSkillByNameErr: repo.ErrSkillNotFound}
	runner, stdout, stderr := testRunner("")

	err := runner.cmdLoadSkill(context.Background(), fake, "bogus")
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
