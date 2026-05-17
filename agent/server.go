package main

import (
	"bytes"
	"log"
	"net/http"
	"os/exec"
	"strings"
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
	return logRequests(allowCORS(mux)), nil
}

type statusRecordingResponseWriter struct {
	http.ResponseWriter
	status    int
	bytes     int
	bodyBytes bytes.Buffer
}

func (w *statusRecordingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecordingResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	if w.bodyBytes.Len() < 1024 {
		_, _ = w.bodyBytes.Write(p[:min(len(p), 1024-w.bodyBytes.Len())])
	}
	return n, err
}

func (w *statusRecordingResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *statusRecordingResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecordingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		if shouldLogRequest(r.URL.Path, status) {
			message := recorder.bodyBytes.String()
			if status >= http.StatusInternalServerError && message != "" {
				log.Printf("http request: method=%s path=%q status=%d bytes=%d remote=%q duration=%s body=%q", r.Method, r.URL.Path, status, recorder.bytes, clientIP(r), time.Since(start).Round(time.Millisecond), message)
			} else {
				log.Printf("http request: method=%s path=%q status=%d bytes=%d remote=%q duration=%s", r.Method, r.URL.Path, status, recorder.bytes, clientIP(r), time.Since(start).Round(time.Millisecond))
			}
		}
	})
}

func shouldLogRequest(path string, status int) bool {
	if status >= http.StatusBadRequest {
		return true
	}
	return strings.HasPrefix(path, "/agent")
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		if comma := strings.IndexByte(forwarded, ','); comma >= 0 {
			return strings.TrimSpace(forwarded[:comma])
		}
		return forwarded
	}
	return r.RemoteAddr
}

func allowCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
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
