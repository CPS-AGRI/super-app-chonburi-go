package database

import (
	"log"
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Seed() {
	var count int64
	DB.Model(&domain.Admin{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("🌱 Seeding initial data...")

	superAdminDeptID := uuid.New()
	supervisorDeptID := uuid.New()
	employeeDeptID := uuid.New()

	departments := []domain.AdminDepartment{
		{
			ID:   superAdminDeptID,
			Name: "superadmin",
			Permissions: []string{
				"MANAGE_CITY_SETTINGS",
			},
		},
		{
			ID:   supervisorDeptID,
			Name: "supervisor",
			Permissions: []string{
				"MANAGE_COMPLAINTS",
				"MANAGE_TAXES",
				"MANAGE_PUBLIC_RELATIONS",
				"VERIFY_CITIZENS",
				"MANAGE_WEATHER_ALERTS",
			},
		},
		{
			ID:   employeeDeptID,
			Name: "employee",
			Permissions: []string{
				"MANAGE_COMPLAINTS",
				"MANAGE_TAXES",
				"MANAGE_PUBLIC_RELATIONS",
				"VERIFY_CITIZENS",
				"MANAGE_WEATHER_ALERTS",
			},
		},
	}

	for i := range departments {
		DB.FirstOrCreate(&departments[i], domain.AdminDepartment{Name: departments[i].Name})
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	admins := []domain.Admin{
		{
			ID:           uuid.New(),
			Email:        "super@admin.com",
			Name:         "Super Admin",
			PhoneNumber:  "0812345678",
			PasswordHash: string(hashedPassword),
			DepartmentID: superAdminDeptID,
		},
		{
			ID:           uuid.New(),
			Email:        "head@admin.com",
			Name:         "Head Admin",
			PhoneNumber:  "0823456789",
			PasswordHash: string(hashedPassword),
			DepartmentID: supervisorDeptID,
		},
		{
			ID:           uuid.New(),
			Email:        "staff@admin.com",
			Name:         "Staff Admin",
			PhoneNumber:  "0834567890",
			PasswordHash: string(hashedPassword),
			DepartmentID: employeeDeptID,
		},
	}

	for i := range admins {
		DB.FirstOrCreate(&admins[i], domain.Admin{Email: admins[i].Email})
	}

	var muniCount int64
	DB.Model(&domain.Municipality{}).Count(&muniCount)
	if muniCount == 0 {
		muni := domain.Municipality{
			ID:            uuid.New(),
			CityNameTh:    "องค์การบริหารส่วนจังหวัดชลบุรี",
			CityNameEn:    "Chonburi Provincial Administrative Organization",
			CityAddressTh: "เลขที่ 999 หมู่ที่ 3 ตำบลเสม็ด อำเภอเมืองชลบุรี จังหวัดชลบุรี 20130",
			CityAddressEn: "999 Moo 3, Samet, Mueang Chonburi, Chonburi 20130",
			CityPhone:     "038-398039",
			CityLogoUrl:   "/logo.webp",
			CityLat:       13.3611,
			CityLng:       100.9847,
			Status:        "active",
			CreatedBy:     "system",
		}
		if err := DB.Create(&muni).Error; err != nil {
			log.Printf("⚠️ Warning: Failed to seed municipality: %v", err)
		} else {
			log.Println("✅ Seeded Municipality successfully")
		}
	}

	log.Println("✅ Seed completed")
}
