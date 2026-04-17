package main

import (
	"fmt"
	"log"
	"time"

	"super-app-chonburi-go/config"
	delivery "super-app-chonburi-go/internal/delivery/http"
	"super-app-chonburi-go/internal/repository"
	"super-app-chonburi-go/internal/usecase"
	"super-app-chonburi-go/pkg/database"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	cfg := config.LoadConfig()

	database.ConnectDB(cfg)
	database.Seed()

	app := fiber.New(fiber.Config{
		AppName:      "Super App Chonburi Backend v3",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	})

	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(logger.New())
	
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	adminRepo := repository.NewAdminRepository(database.DB)
	authUC := usecase.NewAuthUseCase(adminRepo)
	authHandler := delivery.NewAuthHandler(authUC)

	authHandler.RegisterRoutes(app)

	port := ":" + cfg.AppPort
	fmt.Printf("🚀 Go API Backend (Fiber v3) is running on http://localhost%s\n", port)
	
	if err := app.Listen(port); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
