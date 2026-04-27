package main

import (
	"log"
	"time"

	"super-app-chonburi-go/config"
	delivery "super-app-chonburi-go/internal/delivery/http"
	"super-app-chonburi-go/internal/repository"
	"super-app-chonburi-go/internal/usecase"
	"super-app-chonburi-go/pkg/database"
	storage "super-app-chonburi-go/pkg/storage/minio"

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

	minioClient, err := storage.NewClient(cfg.MinIO)
	if err != nil {
		log.Fatalf("Fatal: Failed to initialize MinIO client: %v", err)
	}

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

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Super App Chonburi API is running...")
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))

	adminRepo := repository.NewAdminRepository(database.DB)
	muniRepo := repository.NewMunicipalityRepository(database.DB)
	adminRoleRepo := repository.NewAdminRoleRepository(database.DB)
	permissionRepo := repository.NewSystemPermissionRepository(database.DB)
	deptRepo := repository.NewDepartmentRepository(database.DB)

	authUC := usecase.NewAuthUseCase(adminRepo)
	muniUC := usecase.NewMunicipalityUseCase(muniRepo)
	adminUC := usecase.NewAdminUseCase(adminRepo)
	adminRoleUC := usecase.NewAdminRoleUseCase(adminRoleRepo)
	permissionUC := usecase.NewSystemPermissionUseCase(permissionRepo)
	deptUC := usecase.NewDepartmentUseCase(deptRepo)

	authHandler := delivery.NewAuthHandler(authUC)
	muniHandler := delivery.NewMunicipalityHandler(muniUC)
	uploadHandler := delivery.NewUploadHandler(minioClient, cfg.MinIO)

	api := app.Group("/api/v1")
	authHandler.RegisterRoutes(app)
	muniHandler.RegisterRoutes(app)
	uploadHandler.RegisterRoutes(app)

	port := ":" + cfg.AppPort
	log.Printf("🚀 Server is starting on http://localhost%s", port)

	if err := app.Listen(port); err != nil {
		log.Fatalf("Fatal: Could not start server: %v", err)
	}
}
