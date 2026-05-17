package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

const (
	recipesCLIBinary    = "recipes-cli"
	defaultCLITimeout   = 15 * time.Second
	maxCLITimeout       = 30 * time.Second
	maxCLIOutputBytes   = 64 * 1024
	truncatedOutputNote = "\n...[truncated]\n"
)

type callRecipesCLIArgs struct {
	Args           []string `json:"args" jsonschema:"Arguments to pass to recipes-cli. Use an empty array to run recipes-cli with no arguments and inspect its help output."`
	Stdin          string   `json:"stdin,omitempty" jsonschema:"Optional stdin content. Use this when passing '-' to recipes-cli commands that read JSON from stdin."`
	TimeoutSeconds int      `json:"timeoutSeconds,omitempty" jsonschema:"Optional timeout in seconds. Defaults to 15 and cannot exceed 30."`
}

type callRecipesCLIResult struct {
	Command    string `json:"command"`
	ExitCode   int    `json:"exitCode"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	TimedOut   bool   `json:"timedOut"`
	Successful bool   `json:"successful"`
}

func newRecipesCLITool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "call_recipes_cli",
		Description: "Runs the installed recipes-cli binary with explicit arguments. Use it for recipe list, export, create, patch, import, and schema operations.",
	}, callRecipesCLI)
}

func callRecipesCLI(ctx tool.Context, input callRecipesCLIArgs) (callRecipesCLIResult, error) {
	if err := validateCLIArgs(input.Args); err != nil {
		return callRecipesCLIResult{}, err
	}

	timeout := cliTimeout(input.TimeoutSeconds)
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, recipesCLIBinary, input.Args...)
	if input.Stdin != "" {
		cmd.Stdin = strings.NewReader(input.Stdin)
	}

	var stdout, stderr limitedBuffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := callRecipesCLIResult{
		Command:  strings.Join(append([]string{recipesCLIBinary}, input.Args...), " "),
		ExitCode: exitCode(err),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		TimedOut: errors.Is(runCtx.Err(), context.DeadlineExceeded),
	}
	result.Successful = err == nil

	if err != nil && result.ExitCode == -1 && !result.TimedOut {
		return result, fmt.Errorf("run %s: %w", recipesCLIBinary, err)
	}
	return result, nil
}

func validateCLIArgs(args []string) error {
	for _, arg := range args {
		if strings.ContainsRune(arg, 0) {
			return errors.New("recipes-cli arguments cannot contain NUL bytes")
		}
	}
	return nil
}

func cliTimeout(seconds int) time.Duration {
	if seconds <= 0 {
		return defaultCLITimeout
	}
	timeout := time.Duration(seconds) * time.Second
	if timeout > maxCLITimeout {
		return maxCLITimeout
	}
	return timeout
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

type limitedBuffer struct {
	buf       bytes.Buffer
	truncated bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	accepted := p
	if remaining := maxCLIOutputBytes - b.buf.Len(); remaining <= 0 {
		b.truncated = true
		accepted = nil
	} else if len(p) > remaining {
		b.truncated = true
		accepted = p[:remaining]
	}
	if len(accepted) > 0 {
		if _, err := b.buf.Write(accepted); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	out := b.buf.String()
	if !utf8.ValidString(out) {
		out = strings.ToValidUTF8(out, "\uFFFD")
	}
	if b.truncated {
		out += truncatedOutputNote
	}
	return out
}

var _ io.Writer = (*limitedBuffer)(nil)
