package app

import (
	"context"
	"log"
	"net/http"
	"os"

	"google.golang.org/adk/agent"

	"juancavallotti.com/recipes-agent/internal/config"
	"juancavallotti.com/recipes-agent/internal/copilot"
	"juancavallotti.com/recipes-agent/internal/modelrouter"
	"juancavallotti.com/recipes-agent/internal/server"
)

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

	agentBuilder, ok := registry.AgentBuilder(registry.DefaultAgent)
	if !ok {
		log.Fatalf("model registry: missing default agent builder %q", registry.DefaultAgent)
	}
	imageBuilder, ok := registry.ImageBuilder(registry.DefaultImage)
	if !ok {
		log.Fatalf("model registry: missing default image builder %q", registry.DefaultImage)
	}

	llm, err := agentBuilder(ctx)
	if err != nil {
		log.Fatalf("build default chat model: %v", err)
	}
	imgGen, err := imageBuilder(ctx)
	if err != nil {
		log.Fatalf("build default image generator: %v", err)
	}

	copilot, err := copilot.NewWith(ctx, cfg, llm, imgGen)
	if err != nil {
		log.Fatalf("agent: %v", err)
	}

	handler, err := server.NewHTTPHandler(agent.NewSingleLoader(copilot), cfg, registry)
	if err != nil {
		log.Fatalf("server: %v", err)
	}

	log.Printf("starting recipes agent on %s", cfg.Addr)
	log.Printf("ADK API available under /agent (SSE: /agent/run_sse)")
	if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}
