package app

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"google.golang.org/adk/artifact"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/session"

	"juancavallotti.com/recipes-agent/internal/config"
	"juancavallotti.com/recipes-agent/internal/modelrouter"
	"juancavallotti.com/recipes-agent/internal/observability"
	"juancavallotti.com/recipes-agent/internal/server"
)

const sseWriteTimeout = 120 * time.Second

func Run() {
	config.LoadDotenv()
	cfg := config.Read()
	observability.Init(cfg.LogLevel)
	if cfg.GeminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY is required")
	}

	ctx := context.Background()

	registry, err := modelrouter.BuildRegistry(cfg)
	if err != nil {
		log.Fatalf("model registry: %v", err)
	}

	router := modelrouter.NewRouter(
		registry,
		cfg,
		session.InMemoryService(),
		memory.InMemoryService(),
		artifact.InMemoryService(),
		sseWriteTimeout,
	)

	handler, err := server.NewHTTPHandler(ctx, cfg, router, registry)
	if err != nil {
		log.Fatalf("server: %v", err)
	}

	slog.Info("agent.starting",
		"addr", cfg.Addr,
		"agent_models", len(registry.AgentOptions),
		"image_models", len(registry.ImageOptions),
	)
	if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}
