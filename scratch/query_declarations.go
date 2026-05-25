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

	var declarations []domain.TaxDeclaration
	if err := db.Order("created_at desc").Find(&declarations).Error; err != nil {
		log.Fatalf("failed to query declarations: %v", err)
	}

	for _, d := range declarations {
		fmt.Printf("ID: %s, Ref1: %s, Ref2: %s, CalculatedTax: %f, PaymentStatus: %s, TaxType: %s, Month: %d, Year: %d\n",
			d.ID, d.Ref1, d.Ref2, d.CalculatedTax, d.PaymentStatus, d.TaxType, d.TaxMonth, d.TaxYear)
	}
}
