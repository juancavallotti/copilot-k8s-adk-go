package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/adk/artifact"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/session"

	"juancavallotti.com/recipes-agent/internal/config"
	"juancavallotti.com/recipes-agent/internal/modelrouter"
	"juancavallotti.com/recipes-agent/internal/server"
)

const sseWriteTimeout = 120 * time.Second

func Run() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmicroseconds)

	config.LoadDotenv()
	cfg := config.Read()
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

	log.Printf("starting recipes agent on %s", cfg.Addr)
	log.Printf("ADK API available under /agent (SSE: /agent/run_sse)")
	log.Printf("registered agent models: %d, image models: %d", len(registry.AgentOptions), len(registry.ImageOptions))
	if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}
