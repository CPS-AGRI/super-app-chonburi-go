package database

import (
	"log"

	"super-app-chonburi-go/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func Seed() {
	if DB == nil {
		return
	}

	// 0. WIPE OLD DATA TO ENSURE CLEAN SEED (As requested by user)
	DB.Exec("DELETE FROM admin_departments")
	DB.Exec("DELETE FROM department_permissions")
	DB.Exec("DELETE FROM system_permissions")

	// 1. Seed Permissions (According to the user's provided Image)
	mainModuleID := "COMPLAINT_MODULE"
	permissions := []domain.SystemPermission{
		{ID: "MANAGE_CITY_SETTINGS", NameTh: "จัดการข้อมูลระดับเมือง"},
		
		// Standalone Modules
		{ID: "MODULE_TAX", NameTh: "จัดการภาษี", Description: "จัดการภาษี"},
		{ID: "MODULE_PR", NameTh: "ประชาสัมพันธ์", Description: "ประชาสัมพันธ์"},
		{ID: "MODULE_IDENTITY", NameTh: "การยืนยันตัวตน", Description: "การยืนยันตัวตน"},
		{ID: "MODULE_COMPLAINT_CENTER", NameTh: "ศูนย์ร้องทุกข์ร้องเรียน", Description: "ศูนย์ร้องทุกข์ร้องเรียน"},
		
		// Main Module (ร้องเรียนร้องทุกข์ + Sub Modules)
		{ID: mainModuleID, NameTh: "ร้องเรียนร้องทุกข์", Description: "ร้องเรียนร้องทุกข์"},
		
		// Sub Modules from the image
		{ID: "COMP_FOOD_PERMIT", NameTh: "การขออนุญาตจำหน่ายหรือสะสมอาหาร", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_SPORTS", NameTh: "กีฬาและนันทนาการ", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_TRASH_FEE", NameTh: "ขอชำระค่าธรรมเนียมขยะ", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_LAND_TAX", NameTh: "ขอชำระภาษีที่ดินและสิ่งปลูกสร้าง", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_SIGN_TAX", NameTh: "ขอชำระภาษีป้าย", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_TRAFFIC", NameTh: "ขอบริการจัดจราจร", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_WATER", NameTh: "ขอบริการน้ำอุปโภค - บริโภค", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_ELECTRIC_ADD", NameTh: "ขอเพิ่มไฟฟ้าทางสาธารณะ", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_HEALTH_HAZARD", NameTh: "ขออนุญาตประกอบกิจการที่เป็นอันตรายต่อสุขภาพ", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_CULTURE", NameTh: "งานประเพณี วัฒนธรรม ศาสนา", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_EMERGENCY", NameTh: "แจ้งระงับเหตุต่างๆ", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_CAR_BREAKDOWN", NameTh: "แจ้งเหตุ / ขอความช่วยเหลือรถเสีย", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_REPTILE", NameTh: "แจ้งเหตุจับสัตว์เลื้อยคลาน", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_ROAD_SPILL", NameTh: "แจ้งเหตุสิ่งของตกหล่น รั่วไหล บนถนน", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_ACCIDENT", NameTh: "แจ้งเหตุอุบัติเหตุทางถนน", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_YOUTH", NameTh: "เด็กและเยาวชน", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_TREE_WIRE", NameTh: "ต้นไม้ / กิ่งไม้ ละสายไฟ", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_TREE_OBSTRUCT", NameTh: "ตัดต้นไม้จากเหตุวาตภัย / กีดขวางการจราจร", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_ROAD_DAMAGED", NameTh: "ถนน / หลุมบ่อชำรุด", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_DRAIN_CLOGGED", NameTh: "ท่อระบายน้ำอุดตัน", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_GENERAL", NameTh: "ทั่วไป", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_SIDEWALK", NameTh: "ทางเท้าเสียหาย", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_DISABLED_ALLOWANCE", NameTh: "เบี้ยยังชีพผู้พิการ", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_ELDERLY_ALLOWANCE", NameTh: "เบี้ยยังชีพผู้สูงอายุ", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_TRASH_ISSUE", NameTh: "ปัญหาขยะมูลฝอย", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_DROUGHT", NameTh: "ปัญหาภัยแล้ง", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_ENVIRONMENT", NameTh: "ปัญหาสิ่งแวดล้อม เหตุรำคาญ (กลิ่น แสง เสียง ฝุ่น)", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_AIDS_PATIENT", NameTh: "ผู้ป่วยเอดส์", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_DRAIN_COVER", NameTh: "ฝาท่อชำรุด", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_ELECTRIC_DAMAGED", NameTh: "ไฟฟ้าสาธารณะเสีย", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_SCHOOL_ISSUE", NameTh: "ศูนย์พัฒนาเด็กเล็ก / โรงเรียน", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_POLE_FALLEN", NameTh: "เสาไฟฟ้าเอียง / ล้ม", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_STORM", NameTh: "เหตุวาตภัย", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_FIRE", NameTh: "เหตุอัคคีภัย", ParentID: stringPtr(mainModuleID)},
		{ID: "COMP_FLOOD", NameTh: "เหตุอุทกภัย", ParentID: stringPtr(mainModuleID)},
	}

	for _, p := range permissions {
		DB.FirstOrCreate(&p, domain.SystemPermission{ID: p.ID})
	}

	// 2. Seed Roles
	roles := []domain.AdminRole{
		{Name: "superadmin", IsSuperAdmin: true, Description: stringPtr("ผู้ดูแลระบบสูงสุด")},
		{Name: "supervisor", IsSuperAdmin: false, Description: stringPtr("หัวหน้าหน่วยงาน")},
		{Name: "employee", IsSuperAdmin: false, Description: stringPtr("เจ้าหน้าที่ปฏิบัติงาน")},
	}
	for i := range roles {
		DB.FirstOrCreate(&roles[i], domain.AdminRole{Name: roles[i].Name})
	}
	// Force-update is_superadmin using raw SQL (GORM ignores bool false as zero value)
	DB.Exec("UPDATE admin_roles SET is_superadmin = TRUE WHERE name = 'superadmin'")
	DB.Exec("UPDATE admin_roles SET is_superadmin = FALSE WHERE name != 'superadmin'")

	// 3. Seed Departments
	departments := []domain.Department{
		{
			Name:        "สำนักปลัดเทศบาล",
			Description: "สำนักปลัดเทศบาล",
			Permissions: []domain.SystemPermission{{ID: "COMP_GENERAL"}},
		},
		{
			Name:        "กองคลัง",
			Description: "กองคลัง",
			Permissions: []domain.SystemPermission{{ID: "COMP_TRASH_FEE"}, {ID: "COMP_LAND_TAX"}, {ID: "COMP_SIGN_TAX"}},
		},
		{
			Name:        "กองช่าง",
			Description: "กองช่าง",
			Permissions: []domain.SystemPermission{{ID: mainModuleID}, {ID: "COMP_ROAD_DAMAGED"}, {ID: "COMP_ELECTRIC_DAMAGED"}, {ID: "COMP_DRAIN_CLOGGED"}},
		},
	}

	for i := range departments {
		DB.FirstOrCreate(&departments[i], domain.Department{Name: departments[i].Name})
		if len(departments[i].Permissions) > 0 {
			DB.Model(&departments[i]).Association("Permissions").Replace(departments[i].Permissions)
		}
	}

	// 4. Seed Super Admin
	var superRole domain.AdminRole
	DB.First(&superRole, "name = ?", "superadmin")
	var adminCount int64
	DB.Model(&domain.Admin{}).Count(&adminCount)
	if adminCount == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		superAdmin := domain.Admin{
			Email:        "superadmin@chonburi.go.th",
			Username:     "superadmin",
			Name:         "Super",
			LastName:     "Admin",
			PasswordHash: string(hashedPassword),
			RoleID:       &superRole.ID,
			Status:       "active",
		}
		DB.Create(&superAdmin)
	}

	log.Println("✨ Data Seeding Completed Successfully (According to UI Image)")
}

func stringPtr(s string) *string {
	return &s
}
