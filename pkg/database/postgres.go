package database

import (
	"log"
	"os"
	"time"

	"super-app-chonburi-go/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB(cfg *config.Config) {
	var err error

	// High-Performance & Clean GORM Logger Configuration
	// Default to logger.Warn to silence normal SQL queries and eliminate terminal I/O latency.
	// Queries taking longer than 200ms will be captured as Slow Query alerts.
	logLevel := logger.Warn
	if os.Getenv("DB_DEBUG") == "true" {
		logLevel = logger.Info
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	DB, err = gorm.Open(postgres.Open(cfg.DBDsn), &gorm.Config{
		Logger: gormLogger,
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

	log.Println("[DATABASE] AutoMigrate is completely disabled.")
	log.Println("Database initialized successfully.")
}
