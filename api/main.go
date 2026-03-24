package main

import (
	"learnlang-api/config"
	"learnlang-api/database"
	"learnlang-api/middleware"
	"learnlang-api/routes"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	loc, _ := time.LoadLocation("UTC")
	time.Local = loc
	cfg := config.Load()

	if err := database.Connect(cfg); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := database.ConnectRedis(cfg); err != nil {
		log.Fatal("Failed to connect to redis:", err)
	}

	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if err := database.CreateVectorIndex(); err != nil {
		log.Fatal("Failed to create vector index:", err)
	}

	if err := database.InitUser(cfg); err != nil {
		log.Fatal("Failed to initialize user:", err)
	}

	r := gin.Default()
	r.Use(middleware.CORS())
	routes.SetupRoutes(r, cfg)

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
