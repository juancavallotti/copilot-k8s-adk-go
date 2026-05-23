package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"juancavallotti.com/recipes-cli/internal/commands"
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

	runner := commands.NewRunner(os.Stdin, os.Stdout, os.Stderr, func() (commands.CommandRepo, error) {
		return repo.NewRepo()
	})
	if err := runner.Run(context.Background(), os.Args[1:]); err != nil {
		if errors.Is(err, commands.ErrUsage) {
			os.Exit(2)
		}
		log.Fatal(err)
	}
}
