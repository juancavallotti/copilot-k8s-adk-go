package types

import (
	"encoding/json"
	"time"
)

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	Category     string    `json:"category"`
	Image        string    `json:"image"`
	Photos       []Photo   `json:"photos"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Photo struct {
	ID          string `json:"id,omitempty"`
	ImageBase64 string `json:"image_base64"`
	Featured    bool   `json:"featured"`
}

type Event struct {
	EventID    string    `json:"event_id"`
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at"`
	TraceCount int       `json:"trace_count"`
	UserPrompt string    `json:"user_prompt,omitempty"`
}

type Trace struct {
	ID         string          `json:"id"`
	EventID    string          `json:"event_id"`
	OccurredAt time.Time       `json:"occurred_at"`
	Data       json.RawMessage `json:"data"`
}

type Skill struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RecipeMatch is a search result: a full recipe plus a similarity
// score in [0,1] (cosine similarity, higher = closer match).
type RecipeMatch struct {
	Recipe
	Score float64 `json:"score"`
}

// RecipeHit is the slim form of a search result: enough for an agent
// to decide whether to fetch the full recipe, without dragging photo
// base64 through the context window. Score is the cosine similarity
// in [0,1] of the best-matching chunk; Chunk is that chunk's text.
type RecipeHit struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Chunk string  `json:"chunk"`
	Score float64 `json:"score"`
}

// EventMatch is a search result: a full event plus a similarity
// score in [0,1].
type EventMatch struct {
	Event
	Score float64 `json:"score"`
}
