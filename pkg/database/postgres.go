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

	if cfg.AutoMigrate {
		log.Printf("[ISO-MIGRATION] Starting schema synchronization (env=%s, time=%s)...", cfg.AppEnv, time.Now().UTC().Format(time.RFC3339))
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
			&domain.ModuleUsageLog{},
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
			&domain.CCTVViewLog{},
			&domain.AuditLog{},
		)
		if err != nil {
			log.Fatalf("Fatal: Failed to auto-migrate schema: %v", err)
		}

		// Ensure ISO standard indexes exist idempotently
		DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_activity_date_module ON user_activity_trackings(date, module_id)")
		DB.Exec("CREATE INDEX IF NOT EXISTS idx_module_usage_logs_module ON module_usage_logs(module_code)")
		DB.Exec("CREATE INDEX IF NOT EXISTS idx_module_usage_logs_created_at ON module_usage_logs(created_at)")
		DB.Exec("CREATE INDEX IF NOT EXISTS idx_module_usage_logs_user_id ON module_usage_logs(user_id)")
		DB.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_trace_id ON audit_logs(trace_id)")
		DB.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_request_time ON audit_logs(request_time)")

		log.Printf("[ISO-MIGRATION] Schema synchronized successfully at %s", time.Now().UTC().Format(time.RFC3339))
	} else {
		log.Println("[ISO-MIGRATION] AUTO_MIGRATE=false detected: Skipping AutoMigrate.")
	}
	log.Println("Database initialized successfully.")
}
