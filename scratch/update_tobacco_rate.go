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

	var rate domain.TaxRate
	err = db.Where("tax_type = ?", "tobacco_tax").First(&rate).Error
	if err != nil {
		log.Fatalf("failed to find tobacco_tax rate: %v", err)
	}

	rate.RateValue = 1.0
	if err := db.Save(&rate).Error; err != nil {
		log.Fatalf("failed to update rate: %v", err)
	}

	fmt.Printf("Successfully updated tobacco_tax rate to: %f\n", rate.RateValue)
}
