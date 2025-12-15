package main

import (
	"fmt"
	"os"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/config"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/log"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v", err)
		os.Exit(1)
	}

	log.Init(cfg.LogLevel)
	log.Info("app starting",
		"env", cfg.Env,
		"log_level", cfg.LogLevel,
		"http_port", cfg.HTTPPort,
	)

	server := server.Init(cfg)
	server.Start()
	<-server.Exit
}
