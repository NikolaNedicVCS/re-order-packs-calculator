package main

import (
	"fmt"
	"os"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/config"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/log"
)

func main() {
	fmt.Println("Starting app...")

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v", err)
		os.Exit(1)
	}
	cfg.PrintEnv()

	log.Init(cfg.LogLevel)
}
