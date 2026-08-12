package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"time"
	"user-service/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const databasePingTimeout = 5 * time.Second

type Postgres struct {
	DB    *gorm.DB
	sqlDB *sql.DB
}

func NewPostgres(cfg config.Database) (*Postgres, error) {
	gormDB, err := gorm.Open(postgres.Open(buildPostgresURL(cfg)),
		&gorm.Config{
			DisableAutomaticPing: true,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql database handle: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	ctx, cancel := context.WithTimeout(context.Background(), databasePingTimeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Postgres{
		DB:    gormDB,
		sqlDB: sqlDB,
	}, nil
}

func (p *Postgres) Close() error {
	return p.sqlDB.Close()
}

func buildPostgresURL(cfg config.Database) string {
	connectionURL := &url.URL{
		Scheme: "postgres",
		User: url.UserPassword(
			cfg.User,
			cfg.Password,
		),
		Host: net.JoinHostPort(cfg.Host, cfg.Port),
		Path: "/" + cfg.Name,
	}

	query := connectionURL.Query()
	query.Set("sslmode", cfg.SSLMode)
	connectionURL.RawQuery = query.Encode()
	return connectionURL.String()
}
