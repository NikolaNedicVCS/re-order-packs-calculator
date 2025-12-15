package main

import (
	"fmt"
	"os"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Starting app in %s mode...", cfg.Env)
}
