package modelrouter

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/artifact"
	"google.golang.org/adk/memory"
	adkplugin "google.golang.org/adk/plugin"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/server/adkrest"
	"google.golang.org/adk/session"

	"juancavallotti.com/recipes-agent/internal/config"
	"juancavallotti.com/recipes-agent/internal/copilot"
	"juancavallotti.com/recipes-agent/internal/observability"
)

// Selection identifies the (agent model, image model) combo for one request.
type Selection struct {
	AgentID string
	ImageID string
}

func (s Selection) key() string { return s.AgentID + "|" + s.ImageID }

// Router lazily builds and caches one *adkrest.Server per (agent, image) combo.
// All cached servers share the same session/memory/artifact services so session
// continuity survives model switches.
type Router struct {
	registry        *Registry
	cfg             config.Config
	systemPrompt    string
	sessionService  session.Service
	memoryService   memory.Service
	artifactService artifact.Service
	sseWriteTimeout time.Duration

	cache sync.Map // key string -> *cacheEntry
}

type cacheEntry struct {
	once    sync.Once
	server  *adkrest.Server
	err     error
}

func NewRouter(
	registry *Registry,
	cfg config.Config,
	systemPrompt string,
	sessions session.Service,
	mem memory.Service,
	artifacts artifact.Service,
	sseWriteTimeout time.Duration,
) *Router {
	return &Router{
		registry:        registry,
		cfg:             cfg,
		systemPrompt:    systemPrompt,
		sessionService:  sessions,
		memoryService:   mem,
		artifactService: artifacts,
		sseWriteTimeout: sseWriteTimeout,
	}
}

// Resolve fills in defaults from the registry when either field is empty,
// and validates both IDs against the registered builders.
func (r *Router) Resolve(sel Selection) (Selection, error) {
	if sel.AgentID == "" {
		sel.AgentID = r.registry.DefaultAgent
	}
	if sel.ImageID == "" {
		sel.ImageID = r.registry.DefaultImage
	}
	if _, ok := r.registry.AgentBuilder(sel.AgentID); !ok {
		return sel, fmt.Errorf("unknown agent model %q", sel.AgentID)
	}
	if _, ok := r.registry.ImageBuilder(sel.ImageID); !ok {
		return sel, fmt.Errorf("unknown image model %q", sel.ImageID)
	}
	return sel, nil
}

// HandlerFor returns a cached *adkrest.Server for the given selection.
// The selection must already be validated via Resolve.
func (r *Router) HandlerFor(ctx context.Context, sel Selection) (http.Handler, error) {
	entryAny, _ := r.cache.LoadOrStore(sel.key(), &cacheEntry{})
	entry := entryAny.(*cacheEntry)
	entry.once.Do(func() {
		entry.server, entry.err = r.build(ctx, sel)
	})
	return entry.server, entry.err
}

func (r *Router) build(ctx context.Context, sel Selection) (*adkrest.Server, error) {
	agentBuilder, ok := r.registry.AgentBuilder(sel.AgentID)
	if !ok {
		return nil, fmt.Errorf("unknown agent model %q", sel.AgentID)
	}
	imageBuilder, ok := r.registry.ImageBuilder(sel.ImageID)
	if !ok {
		return nil, fmt.Errorf("unknown image model %q", sel.ImageID)
	}

	llm, err := agentBuilder(ctx)
	if err != nil {
		return nil, fmt.Errorf("build chat model %q: %w", sel.AgentID, err)
	}
	imgGen, err := imageBuilder(ctx)
	if err != nil {
		return nil, fmt.Errorf("build image generator %q: %w", sel.ImageID, err)
	}

	a, err := copilot.NewWith(ctx, r.cfg, r.systemPrompt, llm, imgGen)
	if err != nil {
		return nil, fmt.Errorf("build agent for %s: %w", sel.key(), err)
	}

	eventPlugin, err := observability.NewEventPlugin()
	if err != nil {
		return nil, fmt.Errorf("build observability plugin: %w", err)
	}

	srv, err := adkrest.NewServer(adkrest.ServerConfig{
		AgentLoader:     agent.NewSingleLoader(a),
		SessionService:  r.sessionService,
		MemoryService:   r.memoryService,
		ArtifactService: r.artifactService,
		SSEWriteTimeout: r.sseWriteTimeout,
		PluginConfig: runner.PluginConfig{
			Plugins: []*adkplugin.Plugin{eventPlugin},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create adkrest server for %s: %w", sel.key(), err)
	}
	return srv, nil
}
