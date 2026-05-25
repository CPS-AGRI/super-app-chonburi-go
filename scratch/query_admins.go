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

	var admins []domain.Admin
	if err := db.Preload("Role").Find(&admins).Error; err != nil {
		log.Fatalf("failed to query admins: %v", err)
	}

	for _, a := range admins {
		roleType := ""
		if a.Role != nil {
			roleType = a.Role.Type
		}
		fmt.Printf("ID: %s, Name: %s %s, Email: %s, Role: %s\n", a.ID, a.Name, a.LastName, a.Email, roleType)
	}
}
