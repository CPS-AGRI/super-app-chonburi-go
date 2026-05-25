package main

import (
	"fmt"
	"log"
	"os"
	"super-app-chonburi-go/config"
	"super-app-chonburi-go/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	os.Setenv("DB_DSN", "postgres://uat_admin:C807BI%2B%2BwFizSBH%2FFkNgL7J3qW646npW@122.155.169.235:5489/uat_chonburi?sslmode=disable")
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DBDsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	var rates []domain.TaxRate
	if err := db.Find(&rates).Error; err != nil {
		log.Fatalf("failed to query rates: %v", err)
	}

	for _, r := range rates {
		fmt.Printf("TaxType: %s, NameTH: %s, RateValue: %f, RateUnit: %s, IsActive: %t\n", r.TaxType, r.NameTH, r.RateValue, r.RateUnit, r.IsActive)
	}
}
