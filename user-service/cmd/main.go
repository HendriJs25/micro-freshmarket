package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"
	"user-service/config"
	"user-service/database"
	"user-service/database/seeders"
	"user-service/handler"
	"user-service/repository"
	sessionrepository "user-service/repository/session"
	"user-service/routes"
	"user-service/services"

	"github.com/gin-gonic/gin"
)

const seedTimeout = 30 * time.Second

func main() {
	seed := flag.Bool("seed", false, "seed initial roles and super admin, then exit")
	flag.Parse()

	if err := run(*seed); err != nil {
		log.Fatal(err)
	}
}

func run(seed bool) error {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	postgresDB, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	defer func() {
		if err := postgresDB.Close(); err != nil {
			log.Printf("failed to close postgres connection: %v", err)
		}
	}()

	log.Printf("postgres connection established host=%s port=%s database=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	if seed {
		if err := cfg.ValidateSeed(); err != nil {
			return fmt.Errorf("validate seed configuration: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), seedTimeout)
		defer cancel()

		if err := seeders.Run(ctx, postgresDB.DB, cfg.Seed); err != nil {
			return err
		}

		log.Printf("seed initialized successfully")

		return nil

	}

	redisDB, err := database.NewRedis(cfg.Redis)
	if err != nil {
		return fmt.Errorf("connect to redis: %w", err)
	}

	defer func() {
		if err := redisDB.Close(); err != nil {
			log.Printf("failed to close redis connection: %v", err)
		}
	}()

	log.Printf("redis connection established host=%s port=%s database=%d", cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB)
	sessionRepository := sessionrepository.NewRepository(redisDB.Client)
	repositoryRegistry := repository.NewRegistry(postgresDB.DB)
	serviceRegistry, err := services.NewRegistry(repositoryRegistry, cfg.JWT, sessionRepository)
	if err != nil {
		return fmt.Errorf("initialize services :%w", err)
	}
	handlerRegistry := handler.NewRegistry(serviceRegistry)

	router := gin.Default()
	routeRegistry := routes.NewRegistry(router, handlerRegistry)
	routeRegistry.Register()

	log.Printf(
		"Starting user-service environment=%s address=%s",
		cfg.App.Env,
		cfg.ServerAddress(),
	)

	if err := router.Run(cfg.ServerAddress()); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
	return nil
}
