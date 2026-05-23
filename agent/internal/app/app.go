package app

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"

	"google.golang.org/adk/artifact"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/session"

	"juancavallotti.com/recipes-agent/internal/config"
	"juancavallotti.com/recipes-agent/internal/instruction"
	"juancavallotti.com/recipes-agent/internal/modelrouter"
	"juancavallotti.com/recipes-agent/internal/observability"
	"juancavallotti.com/recipes-agent/internal/server"
	"juancavallotti.com/recipes-agent/internal/skills"
)

const sseWriteTimeout = 120 * time.Second

func Run() {
	config.LoadDotenv()
	cfg := config.Read()

	var traceSinkWriter io.Writer
	if cfg.TracesCLI != "" {
		sink, err := observability.StartTraceSink(cfg.TracesCLI)
		if err != nil {
			log.Printf("trace sink disabled: %v", err)
		} else {
			traceSinkWriter = sink.Writer()
			defer func() {
				if err := sink.Close(); err != nil {
					log.Printf("trace sink close: %v", err)
				}
			}()
		}
	}
	observability.Init(cfg.LogLevel, traceSinkWriter)
	if cfg.GeminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY is required")
	}

	ctx := context.Background()

	template, err := instruction.Load(cfg.InstructionPath)
	if err != nil {
		log.Fatalf("load instruction template: %v", err)
	}
	catalog, err := skills.NewLoader(skills.DefaultCLIBinary).Load(ctx)
	if err != nil {
		log.Fatalf("load skill catalog: %v", err)
	}
	systemPrompt := skills.Render(template, catalog)
	slog.Info("agent.skills_loaded", "count", len(catalog.Skills))

	registry, err := modelrouter.BuildRegistry(cfg)
	if err != nil {
		log.Fatalf("model registry: %v", err)
	}

	router := modelrouter.NewRouter(
		registry,
		cfg,
		systemPrompt,
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
