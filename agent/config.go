package main

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	defaultAddr                       = "localhost:4100"
	defaultModel                      = "gemini-3.1-flash-lite"
	defaultImageModel                 = "gemini-3.1-flash-image-preview"
	defaultImageGenerationConcurrency = 3
	defaultInstructionPath            = "prompts/recipe_copilot.md"
)

type config struct {
	Addr                       string
	Model                      string
	ImageModel                 string
	ImageGenerationConcurrency int
	InstructionPath            string
	GeminiAPIKey               string
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
		Addr:                       os.Getenv("AGENT_ADDR"),
		Model:                      os.Getenv("AGENT_MODEL"),
		ImageModel:                 os.Getenv("AGENT_IMAGE_MODEL"),
		ImageGenerationConcurrency: readBoundedIntEnv("AGENT_IMAGE_GENERATION_CONCURRENCY", defaultImageGenerationConcurrency, maxGeneratedRecipePhotoCount),
		InstructionPath:            os.Getenv("AGENT_INSTRUCTION_PATH"),
		GeminiAPIKey:               os.Getenv("GEMINI_API_KEY"),
	}
	if cfg.Addr == "" {
		cfg.Addr = defaultAddr
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	if cfg.ImageModel == "" {
		cfg.ImageModel = defaultImageModel
	}
	if cfg.InstructionPath == "" {
		cfg.InstructionPath = defaultInstructionPath
	}
	return cfg
}

func readBoundedIntEnv(name string, defaultValue int, maxValue int) int {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		log.Printf("config: invalid %s=%q; using default %d", name, value, defaultValue)
		return defaultValue
	}
	if parsed > maxValue {
		return maxValue
	}
	return parsed
}
