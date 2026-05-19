package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"juancavallotti.com/recipes-agent/internal/modelrouter"
)

// newModelRoutingHandler returns an http.Handler that inspects the JSON body
// of POST /run and POST /run_sse, extracts a top-level "modelContext" object,
// strips it from the forwarded body (ADK rejects unknown fields), and
// delegates to the cached per-combo adkrest.Server returned by the router.
//
// All other paths and methods go to the default server (Google defaults).
func newModelRoutingHandler(router *modelrouter.Router, defaultServer http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !isRunPath(r.URL.Path) {
			defaultServer.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		_ = r.Body.Close()
		if err != nil {
			http.Error(w, "read request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		sel, cleaned, err := extractAndStripModelContext(body)
		if err != nil {
			http.Error(w, "invalid modelContext: "+err.Error(), http.StatusBadRequest)
			return
		}
		resolved, err := router.Resolve(sel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		handler, err := router.HandlerFor(r.Context(), resolved)
		if err != nil {
			http.Error(w, "load model: "+err.Error(), http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(cleaned))
		r.ContentLength = int64(len(cleaned))
		r.Header.Set("Content-Length", strconv.Itoa(len(cleaned)))
		handler.ServeHTTP(w, r)
	})
}

func isRunPath(path string) bool {
	return path == "/run" || path == "/run_sse"
}

// extractAndStripModelContext parses the JSON body, pulls out top-level
// modelContext.{agentModel,imageModel}, and returns the body without that
// field. If modelContext is absent or empty, the returned selection has
// empty IDs which the router treats as "use defaults".
func extractAndStripModelContext(body []byte) (modelrouter.Selection, []byte, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return modelrouter.Selection{}, body, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err != nil {
		return modelrouter.Selection{}, nil, err
	}

	mcRaw, ok := obj["modelContext"]
	if !ok {
		return modelrouter.Selection{}, body, nil
	}
	delete(obj, "modelContext")

	sel := modelrouter.Selection{}
	if mc, ok := mcRaw.(map[string]any); ok {
		sel.AgentID, _ = mc["agentModel"].(string)
		sel.ImageID, _ = mc["imageModel"].(string)
	} else if mcRaw != nil {
		return modelrouter.Selection{}, nil, errors.New("modelContext must be an object")
	}

	cleaned, err := json.Marshal(obj)
	if err != nil {
		return modelrouter.Selection{}, nil, err
	}
	return sel, cleaned, nil
}
