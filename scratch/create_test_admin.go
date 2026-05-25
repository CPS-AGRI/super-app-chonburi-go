package main

import (
	"fmt"
	"log"
	"os"
	"super-app-chonburi-go/config"
	"super-app-chonburi-go/internal/domain"
	"time"

	"golang.org/x/crypto/bcrypt"
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

	// Find super@admin.com to get its role ID
	var superAdmin domain.Admin
	if err := db.Where("email = ?", "super@admin.com").First(&superAdmin).Error; err != nil {
		log.Fatalf("failed to find super@admin.com: %v", err)
	}
	if superAdmin.RoleId == nil {
		log.Fatalf("super@admin.com does not have a RoleId")
	}
	roleID := *superAdmin.RoleId
	fmt.Printf("Found SuperAdmin Role ID from super@admin.com: %s\n", roleID)

	// Hash password "password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	email := "testadmin@example.com"
	var existing domain.Admin
	err = db.Where("email = ?", email).First(&existing).Error
	if err == nil {
		// Update password
		existing.PasswordHash = string(hashedPassword)
		existing.RoleId = &roleID
		if err := db.Save(&existing).Error; err != nil {
			log.Fatalf("failed to update test admin: %v", err)
		}
		fmt.Printf("Updated existing test admin: %s\n", email)
	} else {
		// Create new
		newAdmin := domain.Admin{
			ID:           domain.NewUUID(),
			Name:         "Test",
			LastName:     "Admin",
			Email:        email,
			Phone:        "0812345678",
			Position:     "Tester",
			PasswordHash: string(hashedPassword),
			RoleId:       &roleID,
			CreatedBy:    "system",
			CreatedDate:  time.Now(),
			UpdatedBy:    "system",
			UpdatedDate:  time.Now(),
		}
		if err := db.Create(&newAdmin).Error; err != nil {
			log.Fatalf("failed to create test admin: %v", err)
		}
		fmt.Printf("Created new test admin: %s\n", email)
	}
}
