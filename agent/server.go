package main

import (
	"net/http"
	"os/exec"
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/artifact"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/server/adkrest"
	"google.golang.org/adk/session"
)

func newHTTPHandler(loader agent.Loader, cfg config) (http.Handler, error) {
	restServer, err := adkrest.NewServer(adkrest.ServerConfig{
		AgentLoader:     loader,
		SessionService:  session.InMemoryService(),
		MemoryService:   memory.InMemoryService(),
		ArtifactService: artifact.InMemoryService(),
		SSEWriteTimeout: 120 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/livez", liveness)
	mux.HandleFunc("/readyz", readiness(cfg))
	mux.HandleFunc("/agent", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/agent/", http.StatusTemporaryRedirect)
	})
	mux.Handle("/agent/", http.StripPrefix("/agent", restServer))
	return mux, nil
}

func liveness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func readiness(cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if cfg.GeminiAPIKey == "" {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unready","reason":"missing GEMINI_API_KEY"}`))
			return
		}
		if _, err := exec.LookPath(recipesCLIBinary); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unready","reason":"recipes-cli not found"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}
}
