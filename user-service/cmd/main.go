package main

import (
	"log"
	"user-service/config"
	"user-service/database"
	"user-service/handler"
	"user-service/routes"

	"github.com/gin-gonic/gin"
)

func main() {
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

	router := gin.Default()

	handlerRegistry := handler.NewRegistry()

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
}
