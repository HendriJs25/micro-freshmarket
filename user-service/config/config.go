package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultAppEnv  = "development"
	defaultAppPort = "8090"
)

type Config struct {
	App App
}

type App struct {
	Env  string
	Port string
}

func Load() (*Config, error) {
	cfg := &Config{
		App: App{
			Env:  getEnv("APP_ENV", defaultAppEnv),
			Port: getEnv("APP_PORT", defaultAppPort),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) ServerAddress() string {
	return ":" + c.App.Port
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.App.Env) == "" {
		return fmt.Errorf("APP_ENV must not be empty")
	}

	port, err := strconv.Atoi(c.App.Port)
	if err != nil {
		return fmt.Errorf("APP_PORT must be a number: %w", err)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("APP_PORT must be between 1 and 65535")
	}

	return nil
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	return value
}
