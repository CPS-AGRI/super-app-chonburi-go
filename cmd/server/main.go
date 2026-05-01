package main

import (
	"log"
	"time"

	"super-app-chonburi-go/config"
	delivery "super-app-chonburi-go/internal/delivery/http"
	"super-app-chonburi-go/internal/repository"
	"super-app-chonburi-go/internal/usecase"
	"super-app-chonburi-go/pkg/database"
	"super-app-chonburi-go/internal/delivery/http/middleware"

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
	auditRepo := repository.NewActivityLogRepository(database.DB)
	rtRepo := repository.NewRefreshTokenRepository(database.DB)
	bankRepo := repository.NewMunicipalityBankRepository(database.DB)

	// Apply Global Middleware
	app.Use(middleware.AuditLog(auditRepo))

	authUC := usecase.NewAuthUseCase(adminRepo, rtRepo)
	muniUC := usecase.NewMunicipalityUseCase(muniRepo)
	adminUC := usecase.NewAdminUseCase(adminRepo)
	adminRoleUC := usecase.NewAdminRoleUseCase(adminRoleRepo)
	permissionUC := usecase.NewSystemPermissionUseCase(permissionRepo)
	deptUC := usecase.NewDepartmentUseCase(deptRepo)
	bankUC := usecase.NewMunicipalityBankUseCase(bankRepo)

	authHandler := delivery.NewAuthHandler(authUC)
	muniHandler := delivery.NewMunicipalityHandler(muniUC)
	adminHandler := delivery.NewAdminHandler(adminUC)
	adminRoleHandler := delivery.NewAdminRoleHandler(adminRoleUC)
	permissionHandler := delivery.NewSystemPermissionHandler(permissionUC)
	deptHandler := delivery.NewDepartmentHandler(deptUC)
	bankHandler := delivery.NewMunicipalityBankHandler(bankUC)

	api := app.Group("/api/v1")
	// Protected routes example:
	// api.Use(jwtutil.RequireAuth())
	
	authHandler.RegisterRoutes(app)
	muniHandler.RegisterRoutes(app)
	adminHandler.RegisterRoutes(api)
	adminRoleHandler.RegisterRoutes(api)
	permissionHandler.RegisterRoutes(api)
	deptHandler.RegisterRoutes(api)
	bankHandler.RegisterRoutes(app)

	port := ":" + cfg.AppPort
	log.Printf("🚀 Server is starting on http://localhost%s", port)
	
	if err := app.Listen(port); err != nil {
		log.Fatalf("Fatal: Could not start server: %v", err)
	}
}
