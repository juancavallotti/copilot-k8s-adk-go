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
