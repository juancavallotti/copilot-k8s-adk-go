package main

import (
	"errors"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"juancavallotti.com/recipes-api/handlers"
	"juancavallotti.com/recipes-repo"
)

func loadDotenv() {
	for _, path := range []string{".env", "backend/.env"} {
		if err := godotenv.Load(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			log.Printf("dotenv: load %q: %v", path, err)
		}
	}
}

func main() {
	loadDotenv()

	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = "localhost:4000"
	}

	r, err := repo.NewRepo()
	if err != nil {
		log.Fatalf("repo: %v", err)
	}

	router := gin.Default()
	handlers.New(r).Register(router)

	if err := router.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
