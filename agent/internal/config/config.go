package config

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
	defaultOpenAIModel                = "gpt-5.4-mini"
	defaultOpenAIImageModel           = "gpt-image-1-mini"
	defaultAnthropicModel             = "claude-haiku-4-5"
	DefaultImageGenerationConcurrency = 3
	MaxGeneratedRecipePhotoCount      = 4
	DefaultInstructionPath            = "prompts/recipe_copilot.md"
	DefaultImageOutputDir             = "/tmp/recipe-agent-images"
)

type Config struct {
	Addr                       string
	Model                      string
	ImageModel                 string
	ImageGenerationConcurrency int
	ImageOutputDir             string
	InstructionPath            string
	GeminiAPIKey               string

	OpenAIAPIKey     string
	OpenAIModel      string
	OpenAIImageModel string

	AnthropicAPIKey string
	AnthropicModel  string
}

func LoadDotenv() {
	for _, path := range []string{".env", "agent/.env"} {
		if err := godotenv.Load(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			log.Printf("dotenv: load %q: %v", path, err)
		}
	}
}

func Read() Config {
	cfg := Config{
		Addr:                       os.Getenv("AGENT_ADDR"),
		Model:                      os.Getenv("AGENT_MODEL"),
		ImageModel:                 os.Getenv("AGENT_IMAGE_MODEL"),
		ImageGenerationConcurrency: readBoundedIntEnv("AGENT_IMAGE_GENERATION_CONCURRENCY", DefaultImageGenerationConcurrency, MaxGeneratedRecipePhotoCount),
		ImageOutputDir:             os.Getenv("AGENT_IMAGE_OUTPUT_DIR"),
		InstructionPath:            os.Getenv("AGENT_INSTRUCTION_PATH"),
		GeminiAPIKey:               os.Getenv("GEMINI_API_KEY"),
		OpenAIAPIKey:               os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:                os.Getenv("AGENT_OPENAI_MODEL"),
		OpenAIImageModel:           os.Getenv("AGENT_OPENAI_IMAGE_MODEL"),
		AnthropicAPIKey:            os.Getenv("ANTHROPIC_API_KEY"),
		AnthropicModel:             os.Getenv("AGENT_ANTHROPIC_MODEL"),
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
	if cfg.OpenAIModel == "" {
		cfg.OpenAIModel = defaultOpenAIModel
	}
	if cfg.OpenAIImageModel == "" {
		cfg.OpenAIImageModel = defaultOpenAIImageModel
	}
	if cfg.AnthropicModel == "" {
		cfg.AnthropicModel = defaultAnthropicModel
	}
	if cfg.InstructionPath == "" {
		cfg.InstructionPath = DefaultInstructionPath
	}
	if cfg.ImageOutputDir == "" {
		cfg.ImageOutputDir = DefaultImageOutputDir
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
