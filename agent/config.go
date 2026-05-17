package main

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	defaultAddr  = "localhost:4100"
	defaultModel = "gemini-3.1-flash-lite"
)

type config struct {
	Addr         string
	Model        string
	GeminiAPIKey string
}

func loadDotenv() {
	for _, path := range []string{".env", "agent/.env"} {
		if err := godotenv.Load(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			log.Printf("dotenv: load %q: %v", path, err)
		}
	}
}

func readConfig() config {
	cfg := config{
		Addr:         os.Getenv("AGENT_ADDR"),
		Model:        os.Getenv("AGENT_MODEL"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
	}
	if cfg.Addr == "" {
		cfg.Addr = defaultAddr
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	return cfg
}
