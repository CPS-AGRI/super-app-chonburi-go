package repository

import (
	"strings"
	"super-app-chonburi-go/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) domain.DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) GetBackOffice(filter domain.DashboardFilter) (*domain.DashboardBackOfficeModuleDTO, error) {
	dateNow := time.Now().UTC().Truncate(24 * time.Hour)

	var users []domain.AppUser
	r.db.Preload("Information").Find(&users)

	var allModules []domain.Module
	r.db.Order("sequence ASC, id ASC").Find(&allModules)

	// 1. Demographics & Gender breakdown from real registered users
	maleUsers := 0
	femaleUsers := 0
	var ageGender domain.AgeGenderDTO

	for _, u := range users {
		if u.Information == nil {
			continue
		}

		info := u.Information
		if info.Prefix == "นาย" || info.Prefix == "Mr" || info.Prefix == "Mr." || info.Prefix == "เด็กชาย" {
			maleUsers++
			if info.Birthday != nil {
				age := calculateAge(*info.Birthday, dateNow)
				switch {
				case age >= 20 && age <= 30:
					ageGender.Male.AgeGroup20To30++
				case age >= 31 && age <= 40:
					ageGender.Male.AgeGroup31To40++
				case age >= 41 && age <= 50:
					ageGender.Male.AgeGroup41To50++
				case age >= 51 && age <= 60:
					ageGender.Male.AgeGroup51To60++
				case age >= 61:
					ageGender.Male.AgeGroup61AndAbove++
				}
			}
		} else if info.Prefix == "นาง" || info.Prefix == "นางสาว" || info.Prefix == "Mrs" || info.Prefix == "Miss" || info.Prefix == "เด็กหญิง" {
			femaleUsers++
			if info.Birthday != nil {
				age := calculateAge(*info.Birthday, dateNow)
				switch {
				case age >= 20 && age <= 30:
					ageGender.Female.AgeGroup20To30++
				case age >= 31 && age <= 40:
					ageGender.Female.AgeGroup31To40++
				case age >= 41 && age <= 50:
					ageGender.Female.AgeGroup41To50++
				case age >= 51 && age <= 60:
					ageGender.Female.AgeGroup51To60++
				case age >= 61:
					ageGender.Female.AgeGroup61AndAbove++
				}
			}
		}
	}

	totalUsers := len(users)
	var persentMale, persentFemale float64
	if totalUsers > 0 {
		persentMale = (float64(maleUsers) / float64(totalUsers)) * 100.0
		persentFemale = (float64(femaleUsers) / float64(totalUsers)) * 100.0
	}

	verifiedUsers := 0
	for _, u := range users {
		if u.Information != nil && u.Information.VerificationStatus == "verified" {
			verifiedUsers++
		}
	}

	// 2. Date Filtering Boundaries
	startDate := dateNow.AddDate(0, 0, -27)
	if filter.StartDate != nil {
		startDate = filter.StartDate.UTC()
	}
	endDate := dateNow.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	if filter.EndDate != nil {
		endDate = filter.EndDate.UTC()
		if endDate.Hour() == 0 && endDate.Minute() == 0 && endDate.Second() == 0 {
			endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	// 3. Query audit_logs aggregated by Module in this exact date range
	type ModuleAuditResult struct {
		ModuleGroup string `gorm:"column:module_group"`
		Count       int64  `gorm:"column:c"`
	}
	var auditResults []ModuleAuditResult
	auditQuery := `
		SELECT 
			CASE 
				WHEN path LIKE '/api/v1/complaint%' THEN 'complaint'
				WHEN path LIKE '/api/v1/admin/tax-new%' OR path LIKE '/api/v1/tax%' THEN 'tax'
				WHEN path LIKE '/api/v1/public-relation%' OR path LIKE '/api/v1/modules/%/public-relations%' THEN 'public_relations'
				WHEN path LIKE '/api/v1/verification%' THEN 'user_verify'
				WHEN path LIKE '/api/v1/cctv%' THEN 'cctv'
				WHEN path LIKE '/api/v1/weather%' THEN 'weather'
				WHEN path LIKE '/api/v1/water-level%' OR path LIKE '/api/v1/flood%' THEN 'flood'
				WHEN path LIKE '/api/v1/admin%' OR path LIKE '/api/v1/department%' OR path LIKE '/api/v1/municipality%' THEN 'back_office'
				ELSE 'other'
			END as module_group,
			COUNT(*) as c
		FROM audit_logs
		WHERE request_time >= ? AND request_time <= ?
		GROUP BY module_group
	`
	_ = r.db.Raw(auditQuery, startDate, endDate).Scan(&auditResults).Error

	auditCountMap := make(map[string]int64)
	for _, ar := range auditResults {
		auditCountMap[ar.ModuleGroup] = ar.Count
	}

	// 4. Map real counts to all 9 modules
	var moduleUsages []domain.ModuleUsageDTO
	totalModuleUsage := 0
	for _, m := range allModules {
		count := r.resolveModuleCount(m, auditCountMap, filter)
		totalModuleUsage += count
		moduleUsages = append(moduleUsages, domain.ModuleUsageDTO{
			ModuleId: m.ID,
			ModuleDisplayName: domain.ModuleNameUsageDTO{
				Th: m.NameTh,
				En: m.NameEn,
			},
			ModuleUsageCount: count,
		})
	}

	// 5. Query daily usage activity from audit_logs
	type DailyAuditRow struct {
		Day   time.Time `gorm:"column:d"`
		Count int       `gorm:"column:c"`
	}
	var dailyAuditRows []DailyAuditRow
	dailyQuery := `
		SELECT DATE(request_time) as d, COUNT(*) as c
		FROM audit_logs
		WHERE request_time >= ? AND request_time <= ?
		GROUP BY DATE(request_time)
		ORDER BY d ASC
	`
	_ = r.db.Raw(dailyQuery, startDate, endDate).Scan(&dailyAuditRows).Error

	dailyAuditMap := make(map[string]int)
	for _, dar := range dailyAuditRows {
		k := dar.Day.Format("2006-01-02")
		dailyAuditMap[k] = dar.Count
	}

	// 6. Build daily timeline trend points
	startDay := startDate.Truncate(24 * time.Hour)
	endDay := endDate.Truncate(24 * time.Hour)

	var dailyUsers []domain.DailyUserDTO
	for d := startDay; !d.After(endDay); d = d.Add(24 * time.Hour) {
		endOfDay := d.Add(24 * time.Hour)
		dailyKey := d.Format("2006-01-02")

		daily := domain.DailyUserDTO{
			Date: d,
		}

		// Count citizens registered and verified ON THIS SPECIFIC DAY (Daily, not cumulative)
		for _, u := range users {
			uCreated := u.CreatedDate.UTC()
			if (uCreated.After(d) || uCreated.Equal(d)) && uCreated.Before(endOfDay) {
				daily.TotalDailyUserUsageCount++
			}

			if u.Information != nil && u.Information.VerificationStatus == "verified" {
				if u.Information.VerifiedDate != nil {
					vDate := u.Information.VerifiedDate.UTC()
					if (vDate.After(d) || vDate.Equal(d)) && vDate.Before(endOfDay) {
						daily.TotalDailyVerifiedUserUsageCount++
					}
				} else {
					if (uCreated.After(d) || uCreated.Equal(d)) && uCreated.Before(endOfDay) {
						daily.TotalDailyVerifiedUserUsageCount++
					}
				}
			}
		}

		// Daily actual usage activity frequency from audit_logs
		dailyActivity := dailyAuditMap[dailyKey]
		// Also count new registrations on this day as interaction
		for _, u := range users {
			if u.CreatedDate.Truncate(24 * time.Hour).Equal(d) {
				dailyActivity++
			}
		}

		daily.TotalDailyLoginUserUsageCount = dailyActivity
		dailyUsers = append(dailyUsers, daily)
	}

	result := &domain.DashboardBackOfficeModuleDTO{
		TotalUsers:           totalUsers,
		TotalVerifiedUsers:   verifiedUsers,
		TotalLoginUsersCount: totalModuleUsage,
		PersentUsers: domain.PersentUsersDTO{
			PersentMaleUsers:   persentMale,
			PersentFemaleUsers: persentFemale,
		},
		AgeGender:        ageGender,
		TotalModuleUsage: totalModuleUsage,
		DailyUsers:       dailyUsers,
		ModuleUsage:      moduleUsages,
	}

	return result, nil
}

func (r *dashboardRepository) resolveModuleCount(m domain.Module, auditCounts map[string]int64, filter domain.DashboardFilter) int {
	key := ""
	if m.Key != nil {
		key = strings.ToLower(*m.Key)
	}
	nameTh := strings.ToLower(m.NameTh)
	nameEn := strings.ToLower(m.NameEn)
	identifier := key + " " + nameTh + " " + nameEn

	var count int64 = 0

	if strings.Contains(identifier, "ศูนย์") && (strings.Contains(identifier, "ร้องเรียน") || strings.Contains(identifier, "complaint")) {
		// ศูนย์ร้องทุกข์ร้องเรียน (Complaint Center)
		count = auditCounts["complaint"] / 2
		if count == 0 {
			_ = r.db.Table("module_complaints").Count(&count).Error
		}
	} else if strings.Contains(identifier, "ร้องเรียน") || strings.Contains(identifier, "ร้องทุกข์") || strings.Contains(identifier, "complaint") {
		// ร้องทุกข์ร้องเรียน (Complaint Citizen)
		count = auditCounts["complaint"]
		if count == 0 {
			_ = r.db.Table("module_complaints").Count(&count).Error
		}
	} else if strings.Contains(identifier, "ภาษี") || strings.Contains(identifier, "tax") {
		// จัดการภาษี (Online Tax Payment)
		count = auditCounts["tax"]
		if count == 0 {
			_ = r.db.Table("tax_declarations").Count(&count).Error
		}
	} else if strings.Contains(identifier, "ประชาสัมพันธ์") || strings.Contains(identifier, "ข่าว") || strings.Contains(identifier, "public_relation") || strings.Contains(identifier, "pr") {
		// ประชาสัมพันธ์ (Public Relations)
		prAudit := auditCounts["public_relations"]
		if prAudit > 0 {
			count = prAudit
		} else {
			type sumResult struct {
				Total int64 `gorm:"column:total"`
			}
			var sr sumResult
			_ = r.db.Table("module_public_relation_visitor_count").Select("COALESCE(SUM(count), 0) as total").Scan(&sr).Error
			var newsCount int64
			_ = r.db.Table("module_public_relations").Count(&newsCount).Error
			count = sr.Total + newsCount
		}
	} else if strings.Contains(identifier, "ยืนยัน") || strings.Contains(identifier, "verify") || strings.Contains(identifier, "kyc") {
		// การยืนยันตัวตน (Verify User)
		count = auditCounts["user_verify"]
		if count == 0 {
			_ = r.db.Table("user_informations").Count(&count).Error
		}
	} else if strings.Contains(identifier, "cctv") || strings.Contains(identifier, "กล้อง") {
		// กล้องวงจรปิด (CCTV)
		count = auditCounts["cctv"]
		if count == 0 {
			_ = r.db.Table("module_cctv").Count(&count).Error
		}
	} else if strings.Contains(identifier, "อากาศ") || strings.Contains(identifier, "weather") {
		// สภาพอากาศและพยากรณ์อากาศ (Weather)
		count = auditCounts["weather"]
		if count == 0 {
			var stCount int64
			_ = r.db.Table("weather_station_metadata").Count(&stCount).Error
			count = stCount
		}
	} else if strings.Contains(identifier, "น้ำท่วม") || strings.Contains(identifier, "flood") || strings.Contains(identifier, "river") {
		// เตือนภัยน้ำท่วม (Flood Warning)
		count = auditCounts["flood"]
		if count == 0 {
			var stCount int64
			_ = r.db.Table("river_station_metadata").Count(&stCount).Error
			count = stCount
		}
	} else if strings.Contains(identifier, "เมือง") || strings.Contains(identifier, "back_office") || strings.Contains(identifier, "backoffice") || strings.Contains(identifier, "city") {
		// จัดการเมือง (Back Office)
		count = auditCounts["back_office"]
		if count == 0 {
			var adminCount, deptCount int64
			_ = r.db.Table("admin_users").Count(&adminCount).Error
			_ = r.db.Table("departments").Count(&deptCount).Error
			count = adminCount + deptCount
		}
	} else {
		moduleUUID, parseErr := uuid.Parse(m.ID)
		if parseErr == nil {
			var trackingCount int64
			_ = r.db.Table("user_activity_trackings").Where("module_id = ?", moduleUUID).Select("COALESCE(SUM(view_count), 0)").Scan(&trackingCount).Error
			count = trackingCount
		}
	}

	return int(count)
}

func calculateAge(birthday time.Time, today time.Time) int {
	if birthday.IsZero() {
		return 0
	}
	age := today.Year() - birthday.Year()
	if today.YearDay() < birthday.YearDay() {
		age--
	}
	return age
}

func (r *dashboardRepository) SeedMockData(municipalityId string) error {
	var allModules []domain.Module
	r.db.Find(&allModules)

	if len(allModules) == 0 {
		return nil
	}

	dateNow := time.Now().UTC()

	prefixes := []string{"นาย", "นาง", "นางสาว"}

	for i := 0; i < 100; i++ {
		prefix := prefixes[i%3]

		age := 18 + (i % 52)
		birthYear := dateNow.Year() - age
		birthday := time.Date(birthYear, time.Month((i%12)+1), (i%28)+1, 0, 0, 0, 0, time.UTC)

		status := "active"
		if i%10 == 0 {
			status = "pending"
		}

		name := "ทดสอบ"
		lastName := "ระบบ"

		user := domain.AppUser{
			ID:          uuid.New(),
			PhoneNumber: "08" + uuid.New().String()[:8],
			PinHash:     "dummy",
			CreatedBy:   "system",
			CreatedDate: dateNow.AddDate(0, 0, -(i % 30)),
			UpdatedBy:   "system",
			UpdatedDate: dateNow,
		}

		info := domain.UserInformation{
			UserId:      user.ID,
			Prefix:      prefix,
			Name:        name,
			LastName:    lastName,
			Birthday:    &birthday,
			Status:      status,
			CreatedBy:   "system",
			CreatedDate: user.CreatedDate,
		}

		r.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			if err := tx.Create(&info).Error; err != nil {
				return err
			}
			return nil
		})
	}

	var mockActivities []domain.UserActivityTracking
	for i := 0; i < 30; i++ {
		d := dateNow.AddDate(0, 0, -i).Truncate(24 * time.Hour)

		for _, m := range allModules {
			views := 5 + ((i + len(m.NameEn)) % 45)

			moduleUUID, _ := uuid.Parse(m.ID)
			mockActivities = append(mockActivities, domain.UserActivityTracking{
				ID:        uuid.New(),
				Date:      d,
				ModuleId:  moduleUUID,
				ViewCount: views,
			})
		}
	}
	r.db.Create(&mockActivities)

	return nil
}
