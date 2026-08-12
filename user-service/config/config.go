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

	defaultDatabaseHost         = "localhost"
	defaultDatabasePort         = "5432"
	defaultDatabaseSSLMode      = "disable"
	defaultDatabaseMaxOpenConns = 10
	defaultDatabaseMaxIdleConns = 5
)

type Config struct {
	App      App
	Database Database
}

type App struct {
	Env  string
	Port string
}

type Database struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

func Load() (*Config, error) {
	maxOpenConns, err := getEnvInt("DATABASE_MAX_OPEN_CONNECTIONS", defaultDatabaseMaxOpenConns)
	if err != nil {
		return nil, err
	}

	maxIdleConns, err := getEnvInt("DATABASE_MAX_IDLE_CONNECTIONS", defaultDatabaseMaxIdleConns)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		App: App{
			Env:  getEnv("APP_ENV", defaultAppEnv),
			Port: getEnv("APP_PORT", defaultAppPort),
		},
		Database: Database{
			Host:         getEnv("DATABASE_HOST", defaultDatabaseHost),
			Port:         getEnv("DATABASE_PORT", defaultDatabasePort),
			User:         getEnv("DATABASE_USER", ""),
			Password:     getEnv("DATABASE_PASSWORD", ""),
			Name:         getEnv("DATABASE_NAME", ""),
			SSLMode:      getEnv("DATABASE_SSL_MODE", defaultDatabaseSSLMode),
			MaxOpenConns: maxOpenConns,
			MaxIdleConns: maxIdleConns,
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

	if err := validatePort("APP_PORT", c.App.Port); err != nil {
		return err
	}

	if strings.TrimSpace(c.Database.Host) == "" {
		return fmt.Errorf("DATABASE_HOST must not be empty")
	}

	if err := validatePort("DATABASE_PORT", c.Database.Port); err != nil {
		return err
	}

	if strings.TrimSpace(c.Database.User) == "" {
		return fmt.Errorf("DATABASE_USER must not be empty")
	}

	if strings.TrimSpace(c.Database.Password) == "" {
		return fmt.Errorf("DATABASE_PASSWORD must not be empty")
	}

	if strings.TrimSpace(c.Database.Name) == "" {
		return fmt.Errorf("DATABASE_NAME must not be empty")
	}

	if strings.TrimSpace(c.Database.SSLMode) == "" {
		return fmt.Errorf("DATABASE_SSL_MODE must not be empty")
	}

	if c.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("DATABASE_MAX_OPEN_CONNECTIONS must be greater than 0")
	}

	if c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("DATABASE_MAX_IDLE_CONNECTIONS must not be negative")
	}

	if c.Database.MaxOpenConns < c.Database.MaxIdleConns {
		return fmt.Errorf("DATABASE_MAX_IDLE_CONNECTIONS must not exceed DATABASE_MAX_OPEN_CONNECTIONS")
	}

	return nil
}

func validatePort(key, value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("%s must be a number: %w", key, err)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", key)
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

func getEnvInt(key string, fallback int) (int, error) {
	value, exists := os.LookupEnv(key)

	if !exists {
		return fallback, nil
	}

	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number: %w", key, err)
	}

	return valueInt, nil
}
