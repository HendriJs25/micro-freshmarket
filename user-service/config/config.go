package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAppEnv  = "development"
	defaultAppPort = "8090"

	defaultDatabaseHost         = "localhost"
	defaultDatabasePort         = "5432"
	defaultDatabaseSSLMode      = "disable"
	defaultDatabaseMaxOpenConns = 10
	defaultDatabaseMaxIdleConns = 5

	defaultJWTAccessTokenTTL = 24 * time.Hour
	minimumJWTSecretBytes    = 32
)

type Config struct {
	App      App
	Database Database
	Seed     Seed
	JWT      JWT
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

type Seed struct {
	AdminName     string
	AdminEmail    string
	AdminPassword string
}

type JWT struct {
	SecretKey      string
	Issuer         string
	AccessTokenTTL time.Duration
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

	accessTokenTTL, err := getEnvDuration("JWT_ACCESS_TOKEN_TTL", defaultJWTAccessTokenTTL)
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
		Seed: Seed{
			AdminName:     getEnv("SEED_ADMIN_NAME", ""),
			AdminEmail:    getEnv("SEED_ADMIN_EMAIL", ""),
			AdminPassword: getEnv("SEED_ADMIN_PASSWORD", ""),
		},
		JWT: JWT{
			SecretKey:      getEnv("JWT_SECRET_KEY", ""),
			Issuer:         getEnv("JWT_ISSUER", ""),
			AccessTokenTTL: accessTokenTTL,
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

func (c *Config) ValidateSeed() error {
	if strings.TrimSpace(c.Seed.AdminName) == "" {
		return fmt.Errorf("SEED_ADMIN_NAME should not be empty")
	}

	if strings.TrimSpace(c.Seed.AdminEmail) == "" {
		return fmt.Errorf("SEED_ADMIN_EMAIL should not be empty")
	}

	if strings.TrimSpace(c.Seed.AdminPassword) == "" {
		return fmt.Errorf("SEED_ADMIN_PASSWORD should not be empty")
	}

	if len(c.Seed.AdminPassword) < 8 {
		return fmt.Errorf("SEED_ADMIN_PASSWORD should not be less than 8")
	}

	if len(c.Seed.AdminPassword) > 72 {
		return fmt.Errorf("SEED_ADMIN_PASSWORD should not be more than 72")
	}

	return nil
}

func (j JWT) Validate() error {
	if len([]byte(j.SecretKey)) < minimumJWTSecretBytes {
		return fmt.Errorf("JWT_SECRET_KEY must be at least % bytes", minimumJWTSecretBytes)
	}

	if strings.TrimSpace(j.Issuer) == "" {
		return fmt.Errorf("JWT_ISSUER must not be empty")
	}

	if j.AccessTokenTTL <= 0 {
		return fmt.Errorf("JWT_ACCESS_TOKEN_TTL must be greater than zero")
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

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback, nil
	}

	result, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}

	return result, nil
}
