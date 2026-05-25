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

	// Direct updates
	if err := db.Debug().Model(&domain.TaxRate{}).Where("tax_type = ?", "hotel_fee").Update("rate_value", 0.5).Error; err != nil {
		log.Printf("Failed to update hotel_fee rate: %v", err)
	} else {
		fmt.Println("Successfully updated hotel_fee rate to 0.5")
	}

	if err := db.Debug().Model(&domain.TaxRate{}).Where("tax_type = ?", "oil_gas_tax").Update("rate_value", 4.54).Error; err != nil {
		log.Printf("Failed to update oil_gas_tax rate: %v", err)
	} else {
		fmt.Println("Successfully updated oil_gas_tax rate to 4.54")
	}

	if err := db.Debug().Model(&domain.TaxRate{}).Where("tax_type = ?", "tobacco_tax").Update("rate_value", 1.0).Error; err != nil {
		log.Printf("Failed to update tobacco_tax rate: %v", err)
	} else {
		fmt.Println("Successfully updated tobacco_tax rate to 1.0")
	}

	fmt.Println("All rate updates completed.")
}
