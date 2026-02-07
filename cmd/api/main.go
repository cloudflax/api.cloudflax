package main

import (
	"log"

	"github.com/cloudflax/api.cloudflax/internal/app"
	"github.com/cloudflax/api.cloudflax/internal/config"
	"github.com/cloudflax/api.cloudflax/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuración inválida: %v", err)
	}

	if err := db.Init(cfg); err != nil {
		log.Fatalf("Base de datos: %v", err)
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
