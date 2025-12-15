package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env string
}

// Load config from .env if it exists. If it doesn't, fall back to real env vars.
// Env vars keep precedence (godotenv.Load does not override existing env values).
func Load() (Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load(".env")
		if err != nil {
			return Config{}, fmt.Errorf("failed to load .env: %w", err)
		}
	}

	cfg := Config{
		Env: getEnv("APP_ENV", "dev"),
	}

	if err := validate(cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func validate(cfg Config) error {
	var errs []string

	if strings.TrimSpace(cfg.Env) == "" {
		errs = append(errs, "APP_ENV is required")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
