package instruction

import (
	"strings"
	"testing"

	"juancavallotti.com/recipes-agent/internal/config"
)

func TestLoadInstructionFromAgentWorkingDirectory(t *testing.T) {
	t.Chdir("../..")

	instruction, err := Load(config.DefaultInstructionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !strings.Contains(instruction, "You are a copilot for the recipe application.") {
		t.Fatalf("instruction missing expected content: %q", instruction)
	}
}

func TestLoadInstructionFromRepoRootWorkingDirectory(t *testing.T) {
	t.Chdir("../../..")

	instruction, err := Load(config.DefaultInstructionPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !strings.Contains(instruction, "{{SKILL_CATALOG}}") || !strings.Contains(instruction, "{{CLI_HELP}}") {
		t.Fatalf("instruction template missing expected placeholders: %q", instruction)
	}
}
