package database

import (
	"log"
	"super-app-chonburi-go/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

func Seed() {
	var count int64
	DB.Model(&domain.Admin{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("🌱 Seeding...")

	departments := []domain.AdminDepartment{
		{
			Name: "Super Administration",
			Permissions: []string{
				"MANAGE_COMPLAINTS",
				"MANAGE_TAXES",
				"MANAGE_CITY_SETTINGS",
				"MANAGE_PUBLIC_RELATIONS",
				"VERIFY_CITIZENS",
				"MANAGE_WEATHER_ALERTS",
			},
		},
		{
			Name: "Supervisor Team",
			Permissions: []string{
				"MANAGE_COMPLAINTS",
				"MANAGE_PUBLIC_RELATIONS",
			},
		},
		{
			Name: "General Staff",
			Permissions: []string{
				"MANAGE_COMPLAINTS",
			},
		},
	}

	for i := range departments {
		DB.FirstOrCreate(&departments[i], domain.AdminDepartment{Name: departments[i].Name})
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	admins := []domain.Admin{
		{
			Email:        "super@admin.com",
			Name:         "Super Admin",
			PhoneNumber:  "0812345678",
			PasswordHash: string(hashedPassword),
			DepartmentID: departments[0].ID,
		},
		{
			Email:        "head@admin.com",
			Name:         "Head Admin",
			PhoneNumber:  "0823456789",
			PasswordHash: string(hashedPassword),
			DepartmentID: departments[1].ID,
		},
		{
			Email:        "staff@admin.com",
			Name:         "Staff Admin",
			PhoneNumber:  "0834567890",
			PasswordHash: string(hashedPassword),
			DepartmentID: departments[2].ID,
		},
	}

	for i := range admins {
		DB.FirstOrCreate(&admins[i], domain.Admin{Email: admins[i].Email})
	}

	log.Println("✅ Seeded")
}
