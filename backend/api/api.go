package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"juancavallotti.com/recipes-api/handlers"
	"juancavallotti.com/recipes-repo"
)

// shutdownTimeout caps the time we wait for in-flight HTTP requests
// and async embedding goroutines to drain after a shutdown signal.
// 60s matches the per-embedding timeout in the dbops packages so the
// last in-flight write has a full window to finish.
const shutdownTimeout = 60 * time.Second

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

	r, err := repo.NewRepo()
	if err != nil {
		slog.Error("api.repo_init_failed", "err", err)
		os.Exit(1)
	}

	router := gin.New()
	router.Use(slogRequestLogger(), slogRecovery())
	handlers.New(r).Register(router)

	server := &http.Server{
		Addr:    addr,
		Handler: router.Handler(),
	}

	// Run the HTTP server in a goroutine so main can wait for either
	// a fatal listen error or a shutdown signal.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("api.starting", "addr", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		slog.Info("api.shutdown_signal", "sig", sig.String())
	case err := <-serverErr:
		if err != nil {
			slog.Error("api.server_failed", "err", err)
			_ = r.Close()
			os.Exit(1)
		}
	}

	// Phase 1: stop accepting new connections, wait for in-flight
	// requests to return.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("api.http_shutdown_failed", "err", err)
	}

	// Phase 2: drain async embedding goroutines and close the DB pool.
	// Repo.Close handles both — Wait on each store, then pool.Close.
	if err := r.Close(); err != nil {
		slog.Error("api.repo_close_failed", "err", err)
	}

	slog.Info("api.shutdown_complete")
}
