package database

import (
	"log"
	"time"

	"super-app-chonburi-go/config"
	"super-app-chonburi-go/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB(cfg *config.Config) {
	var err error

	gormLogLevel := logger.Error
	if cfg.AppEnv == "development" {
		gormLogLevel = logger.Info
	}

	DB, err = gorm.Open(postgres.Open(cfg.DBDsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})

	if err != nil {
		log.Fatalf("Fatal: Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Fatal: Failed to extract underlying sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	if cfg.AppEnv != "production" {
		log.Println("Migrating MueangSmart schema (snake_case)...")
		err = DB.AutoMigrate(
			&domain.AdminRole{},
			&domain.Admin{},
			&domain.AdminRefreshToken{},
			&domain.Department{},
			&domain.Module{},
			&domain.ModuleType{},
			&domain.DepartmentModule{},
			&domain.DepartmentModuleModuleType{},
			&domain.Complaint{},
			&domain.ComplaintImage{},
			&domain.ComplaintActivity{},
			&domain.ComplaintActivityImage{},
			&domain.ComplaintRatingHistory{},
			&domain.Municipality{},
			&domain.MunicipalityBank{},
			&domain.MunicipalityWorkSchedule{},
			&domain.AppUser{},
			&domain.UserInformation{},
			&domain.UserOauthAccount{},
			&domain.UserActivityTracking{},
			&domain.PublicRelation{},
			&domain.PublicRelationVisitorCount{},
			&domain.PublicRelationNotification{},
			&domain.PublicRelationLike{},
			&domain.PublicRelationImage{},
			&domain.PublicRelationComment{},
			&domain.MunicipalityWelcomeScreen{},
			&domain.ModuleNotification{},
			&domain.ModuleUserNotification{},
			&domain.ModuleDeviceToken{},
			&domain.UserFCMToken{},
			&domain.TaxRate{},
			&domain.TaxBusiness{},
			&domain.TaxDeclaration{},
			&domain.BankReconciliationBatch{},
			&domain.BankReconciliationRecord{},
			&domain.ElaasDailySummary{},
			&domain.CCTV{},
			&domain.CCTVRequest{},
		)
		if err != nil {
			log.Fatalf("Fatal: Failed to auto-migrate: %v", err)
		}
	} else {
		log.Println("Production environment detected: Skipping AutoMigrate.")
	}
	log.Println("Database initialized successfully.")
}
