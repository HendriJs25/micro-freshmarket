package main

import (
	"log"
	"user-service/config"
	"user-service/handler"
	"user-service/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

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
