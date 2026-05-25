package main

import (
	"encoding/json"
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

	var d domain.TaxDeclaration
	if err := db.Preload("Business").First(&d, "id = ?", "065f18d2-5500-415c-bf3c-5a67f72c5a92").Error; err != nil {
		log.Fatalf("failed to query declaration: %v", err)
	}

	bz, _ := json.MarshalIndent(d, "", "  ")
	fmt.Println(string(bz))
}
