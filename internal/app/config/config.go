package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env      string
	LogLevel string
	HTTPPort string
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
		Env:      getEnv("APP_ENV", "dev"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
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
	if strings.TrimSpace(cfg.LogLevel) == "" {
		errs = append(errs, "LOG_LEVEL is required")
	}
	if strings.TrimSpace(cfg.HTTPPort) == "" {
		errs = append(errs, "HTTP_PORT is required")
	} else if err := validateHTTPPort(cfg.HTTPPort); err != nil {
		errs = append(errs, "HTTP_PORT is invalid: "+err.Error())
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

func validateHTTPPort(port string) error {
	p, err := strconv.Atoi(strings.TrimSpace(port))
	if err != nil {
		return fmt.Errorf("must be a number, provided: %s", port)
	}
	if p < 1 || p > 65535 {
		return fmt.Errorf("must be between 1 and 65535, provided: %d", p)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
