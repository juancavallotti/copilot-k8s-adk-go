package main

import (
	"strings"
	"testing"
)

func TestLoadInstructionFromAgentWorkingDirectory(t *testing.T) {
	instruction, err := loadInstruction(defaultInstructionPath)
	if err != nil {
		t.Fatalf("loadInstruction() error = %v", err)
	}
	if !strings.Contains(instruction, "You are a copilot for the recipe application.") {
		t.Fatalf("instruction missing expected content: %q", instruction)
	}
}

func TestLoadInstructionFromRepoRootWorkingDirectory(t *testing.T) {
	t.Chdir("..")

	instruction, err := loadInstruction(defaultInstructionPath)
	if err != nil {
		t.Fatalf("loadInstruction() error = %v", err)
	}
	if !strings.Contains(instruction, "generate_recipe_photos") {
		t.Fatalf("instruction missing expected tool guidance: %q", instruction)
	}
}
