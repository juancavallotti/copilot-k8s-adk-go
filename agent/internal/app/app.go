package app

import (
	"context"
	"log"
	"net/http"
	"os"

	"google.golang.org/adk/agent"

	"juancavallotti.com/recipes-agent/internal/config"
	"juancavallotti.com/recipes-agent/internal/copilot"
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
	copilot, err := copilot.New(ctx, cfg)
	if err != nil {
		log.Fatalf("agent: %v", err)
	}

	handler, err := server.NewHTTPHandler(agent.NewSingleLoader(copilot), cfg)
	if err != nil {
		log.Fatalf("server: %v", err)
	}

	log.Printf("starting recipes agent on %s", cfg.Addr)
	log.Printf("ADK API available under /agent (SSE: /agent/run_sse)")
	if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}
