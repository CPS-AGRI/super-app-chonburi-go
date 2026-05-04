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

	DB, err = gorm.Open(postgres.Open(cfg.DBDsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
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

	err = DB.AutoMigrate(
		&domain.AdminRole{},
		&domain.Admin{},
		&domain.Department{},
		&domain.SystemPermission{},
		&domain.Municipality{},
		&domain.MunicipalityBank{},
		&domain.CityShift{},
		&domain.ActivityLog{},
		&domain.AdminRefreshToken{},
		&domain.Complaint{},
		&domain.ComplaintUserInformation{},
		&domain.ComplaintImage{},
		&domain.ComplaintActivity{},
		&domain.ComplaintActivityImage{},
	)
	if err != nil {
		log.Fatalf("Fatal: Failed to auto-migrate: %v", err)
	}

	log.Println("Database Connected & Migrated")
}
