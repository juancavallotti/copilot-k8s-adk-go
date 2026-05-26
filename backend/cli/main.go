package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"juancavallotti.com/recipes-cli/internal/commands"
)

type dotenvLoadError struct {
	path string
	err  error
}

func loadDotenv() []dotenvLoadError {
	var errs []dotenvLoadError
	for _, path := range []string{".env", "backend/.env"} {
		if err := godotenv.Load(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			errs = append(errs, dotenvLoadError{path: path, err: err})
		}
	}
	return errs
}

func main() {
	initSlog(os.Stderr, os.Getenv("BACKEND_LOG_LEVEL"))
	dotenvErrs := loadDotenv()
	initSlog(os.Stderr, os.Getenv("BACKEND_LOG_LEVEL"))
	for _, loadErr := range dotenvErrs {
		slog.Warn("dotenv.load_failed", "path", loadErr.path, "err", loadErr.err)
	}

	runner := commands.NewRunnerWithLogger(os.Stdin, os.Stdout, os.Stderr, slog.Default())
	if err := runner.Run(context.Background(), os.Args[1:]); err != nil {
		if errors.Is(err, commands.ErrUsage) {
			os.Exit(2)
		}
		slog.Error("cli.command_failed", "err", err)
		os.Exit(1)
	}
}
