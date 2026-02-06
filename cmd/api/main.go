package main

import (
	"log"
	"os"

	"github.com/cloudflax/api.cloudflax/internal/app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Run(port); err != nil {
		log.Fatal(err)
	}
}
