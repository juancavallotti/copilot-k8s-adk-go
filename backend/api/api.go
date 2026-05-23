package main

import (
	"errors"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"juancavallotti.com/recipes-api/handlers"
	"juancavallotti.com/recipes-repo"
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
	initSlog(os.Stdout, os.Getenv("BACKEND_LOG_LEVEL"))
	dotenvErrs := loadDotenv()
	initSlog(os.Stdout, os.Getenv("BACKEND_LOG_LEVEL"))
	for _, loadErr := range dotenvErrs {
		slog.Warn("dotenv.load_failed", "path", loadErr.path, "err", loadErr.err)
	}

	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = "localhost:4000"
	}

	slog.Info("api.starting", "addr", addr)
	r, err := repo.NewRepo()
	if err != nil {
		slog.Error("api.repo_init_failed", "err", err)
		os.Exit(1)
	}

	router := gin.New()
	router.Use(slogRequestLogger(), slogRecovery())
	handlers.New(r).Register(router)

	if err := router.Run(addr); err != nil {
		slog.Error("api.server_failed", "err", err)
		os.Exit(1)
	}
}
