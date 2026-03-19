package main

import (
	"flag"
	"log"

	"prod-pobeda-2026/internal/app"
	"prod-pobeda-2026/internal/config"
)

// @title           BrandRadar API
// @version         1.0.0
// @description     API системы мониторинга репутации брендов BrandRadar
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         https http
// @produce         json
// @consume         json
func main() {
	configPath := flag.String("config", "config/config.yml", "path to config file")
	flag.Parse()

	if err := config.Init(*configPath); err != nil {
		log.Fatalf("Failed to init config: %v", err)
	}

	app.Run(*configPath)
}
