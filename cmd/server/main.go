package main

import (
	"log"
	"time"

	"super-app-chonburi-go/config"
	delivery "super-app-chonburi-go/internal/delivery/http"
	"super-app-chonburi-go/internal/repository"
	"super-app-chonburi-go/internal/usecase"
	"super-app-chonburi-go/pkg/database"
	"super-app-chonburi-go/pkg/mail"
	"super-app-chonburi-go/pkg/storage"
	minioStorage "super-app-chonburi-go/pkg/storage/minio"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func main() {
	cfg := config.LoadConfig()

	database.ConnectDB(cfg)

	minioClient, err := minioStorage.NewClient(cfg.MinIO)
	if err != nil {
		log.Fatalf("Fatal: Failed to initialize MinIO client: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "Super App Chonburi Backend (MueangSmart)",
		BodyLimit:    100 * 1024 * 1024, // 100MB
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	})

	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(logger.New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Super App Chonburi API (MueangSmart Core) is running...")
	})

	app.Get("/uploads/*", static.New("./uploads"))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:5173"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))

	adminRepo := repository.NewAdminRepository(database.DB)
	adminRoleRepo := repository.NewAdminRoleRepository(database.DB)
	deptRepo := repository.NewDepartmentRepository(database.DB)
	rtRepo := repository.NewRefreshTokenRepository(database.DB)
	complaintRepo := repository.NewComplaintRepository(database.DB)
	moduleRepo := repository.NewModuleRepository(database.DB)
	moduleTypeRepo := repository.NewModuleTypeRepository(database.DB)
	muniRepo := repository.NewMunicipalityRepository(database.DB)
	muniBankRepo := repository.NewMunicipalityBankRepository(database.DB)
	muniWorkScheduleRepo := repository.NewMunicipalityWorkScheduleRepository(database.DB)
	dashboardRepo := repository.NewDashboardRepository(database.DB)
	taxRepo := repository.NewTaxRepository(database.DB)
	taxNewRepo := repository.NewTaxNewRepository(database.DB)
	publicRelationRepo := repository.NewPublicRelationRepository(database.DB)
	verificationRepo := repository.NewVerificationRepository(database.DB)

	// Providers
	emailSender := mail.NewSMTPEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPEmail, cfg.SMTPPassword)
	storageProvider := storage.NewLocalStorage("./uploads", "http://localhost:"+cfg.AppPort+"/uploads")

	// UseCases
	authUC := usecase.NewAuthUseCase(adminRepo, rtRepo)
	adminUC := usecase.NewAdminUseCase(adminRepo)
	adminRoleUC := usecase.NewAdminRoleUseCase(adminRoleRepo)
	deptUC := usecase.NewDepartmentUseCase(deptRepo)
	complaintUC := usecase.NewComplaintUseCase(complaintRepo, adminRepo, muniRepo)
	moduleUC := usecase.NewModuleUseCase(moduleRepo, adminRepo)
	moduleTypeUC := usecase.NewModuleTypeUseCase(moduleTypeRepo)
	muniUC := usecase.NewMunicipalityUseCase(muniRepo)
	muniBankUC := usecase.NewMunicipalityBankUseCase(muniBankRepo)
	muniWorkScheduleUC := usecase.NewMunicipalityWorkScheduleUseCase(muniWorkScheduleRepo)
	dashboardUC := usecase.NewDashboardUseCase(dashboardRepo)
	taxUC := usecase.NewTaxUseCase(taxRepo)
	taxNewUC := usecase.NewTaxNewUseCase(taxNewRepo, emailSender, "")
	publicRelationUC := usecase.NewPublicRelationUseCase(publicRelationRepo, adminRepo)
	verificationUC := usecase.NewVerificationUseCase(verificationRepo)

	// Handlers
	authHandler := delivery.NewAuthHandler(authUC)
	uploadHandler := delivery.NewUploadHandler(minioClient, cfg.MinIO)
	adminHandler := delivery.NewAdminHandler(adminUC)
	adminRoleHandler := delivery.NewAdminRoleHandler(adminRoleUC)
	deptHandler := delivery.NewDepartmentHandler(deptUC)
	complaintHandler := delivery.NewComplaintHandler(complaintUC, deptUC)
	moduleHandler := delivery.NewModuleHandler(moduleUC)
	moduleTypeHandler := delivery.NewModuleTypeHandler(moduleTypeUC)
	muniHandler := delivery.NewMunicipalityHandler(muniUC)
	muniBankHandler := delivery.NewMunicipalityBankHandler(muniBankUC)
	muniWorkScheduleHandler := delivery.NewMunicipalityWorkScheduleHandler(muniWorkScheduleUC)
	dashboardHandler := delivery.NewDashboardHandler(dashboardUC)
	taxHandler := delivery.NewTaxHandler(taxUC)
	taxNewHandler := delivery.NewTaxNewHandler(taxNewUC, storageProvider)
	publicRelationHandler := delivery.NewPublicRelationHandler(publicRelationUC)
	verificationHandler := delivery.NewVerificationHandler(verificationUC)

	api := app.Group("/api/v1")
	
	// Register Routes
	authHandler.RegisterRoutes(api)
	uploadHandler.RegisterRoutes(app)
	adminHandler.RegisterRoutes(api)
	adminRoleHandler.RegisterRoutes(api)
	deptHandler.RegisterRoutes(api)
	complaintHandler.RegisterRoutes(api)
	moduleHandler.RegisterRoutes(api)
	moduleTypeHandler.RegisterRoutes(api)
	muniBankHandler.RegisterRoutes(api)
	muniWorkScheduleHandler.RegisterRoutes(api)
	muniHandler.RegisterRoutes(api)
	dashboardHandler.RegisterRoutes(api)
	taxHandler.RegisterRoutes(api)
	taxNewHandler.RegisterRoutes(api)
	publicRelationHandler.RegisterRoutes(api)
	publicRelationHandler.RegisterGlobalRoutes(api)
	verificationHandler.RegisterRoutes(api)

	port := ":" + cfg.AppPort
	log.Printf("🚀 Server is starting on http://localhost%s", port)

	if err := app.Listen(port); err != nil {
		log.Fatalf("Fatal: Could not start server: %v", err)
	}
}
