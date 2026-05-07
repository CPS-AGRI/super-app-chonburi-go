package repository

import (
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
	r.db.Find(&users)

	var allModules []domain.Module
	r.db.Find(&allModules)

	// User Activity Queries
	activityQuery := r.db.Model(&domain.UserActivityTracking{})
	if filter.StartDate != nil {
		activityQuery = activityQuery.Where("date >= ?", filter.StartDate.Truncate(24*time.Hour))
	}
	if filter.EndDate != nil {
		activityQuery = activityQuery.Where("date <= ?", filter.EndDate.Truncate(24*time.Hour))
	}

	var trackings []domain.UserActivityTracking
	activityQuery.Find(&trackings)

	// Calculate overall usage
	totalLoginUsersCount := 0
	for _, t := range trackings {
		totalLoginUsersCount += t.ViewCount
	}

	// Calculate gender and age
	maleUsers := 0
	femaleUsers := 0
	var ageGender domain.AgeGenderDTO

	for _, u := range users {
		if u.Prefix != nil && (*u.Prefix == "นาย" || *u.Prefix == "Mr" || *u.Prefix == "Mr." || *u.Prefix == "เด็กชาย") {
			maleUsers++
			if u.Birthday != nil {
				age := calculateAge(*u.Birthday, dateNow)
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
		} else if u.Prefix != nil && (*u.Prefix == "นาง" || *u.Prefix == "นางสาว" || *u.Prefix == "Mrs" || *u.Prefix == "Miss" || *u.Prefix == "เด็กหญิง") {
			femaleUsers++
			if u.Birthday != nil {
				age := calculateAge(*u.Birthday, dateNow)
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

	// Daily calculations
	startDate := dateNow
	if filter.StartDate != nil {
		startDate = filter.StartDate.Truncate(24 * time.Hour)
	}
	endDate := dateNow
	if filter.EndDate != nil {
		endDate = filter.EndDate.Truncate(24 * time.Hour)
	}

	var dailyUsers []domain.DailyUserDTO
	for d := startDate; !d.After(endDate); d = d.Add(24 * time.Hour) {
		daily := domain.DailyUserDTO{
			Date: d,
		}

		// Count users created on this day
		for _, u := range users {
			if u.CreatedDate.Truncate(24 * time.Hour).Equal(d) {
				daily.TotalDailyUserUsageCount++
				if u.Status == domain.AppUserStatusActive {
					daily.TotalDailyVerifiedUserUsageCount++
				}
			}
		}

		// Count logins/views on this day
		for _, t := range trackings {
			if t.Date.Truncate(24 * time.Hour).Equal(d) {
				daily.TotalDailyLoginUserUsageCount += t.ViewCount
			}
		}

		dailyUsers = append(dailyUsers, daily)
	}

	// Module calculations
	var moduleUsages []domain.ModuleUsageDTO
	totalModuleUsage := 0
	for _, m := range allModules {
		count := 0
		for _, t := range trackings {
			moduleUUID, _ := uuid.Parse(m.ID)
			if t.ModuleId == moduleUUID {
				count += t.ViewCount
			}
		}
		
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

	// Count verified users overall
	verifiedUsers := 0
	for _, u := range users {
		if u.Status == domain.AppUserStatusActive {
			verifiedUsers++
		}
	}

	result := &domain.DashboardBackOfficeModuleDTO{
		TotalUsers:           totalUsers,
		TotalVerifiedUsers:   verifiedUsers,
		TotalLoginUsersCount: totalLoginUsersCount,
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
		return nil // No modules to seed activity for
	}

	dateNow := time.Now().UTC()

	// 1. Generate 100 AppUsers
	var mockUsers []domain.AppUser
	prefixes := []string{"นาย", "นาง", "นางสาว"}
	
	for i := 0; i < 100; i++ {
		// Randomize prefix
		prefix := prefixes[i%3]
		p := &prefix
		
		// Randomize age between 18 and 70
		age := 18 + (i % 52)
		birthYear := dateNow.Year() - age
		birthday := time.Date(birthYear, time.Month((i%12)+1), (i%28)+1, 0, 0, 0, 0, time.UTC)
		b := &birthday

		status := domain.AppUserStatusActive
		if i%10 == 0 {
			status = domain.AppUserStatusPending
		}

		name := "ทดสอบ"
		lastName := "ระบบ"

		mockUsers = append(mockUsers, domain.AppUser{
			ID:             uuid.New(),
			Prefix:         p,
			Name:           &name,
			LastName:       &lastName,
			Birthday:       b,
			Status:         status,
			PhoneNumber:    "08" + uuid.New().String()[:8],
			PinHash:        "dummy",
			CreatedBy:      "system",
			CreatedDate:    dateNow.AddDate(0, 0, -(i % 30)),
			UpdatedBy:      "system",
			UpdatedDate:    dateNow,
		})
	}
	r.db.Create(&mockUsers)

	// 2. Generate Activity Tracking for last 30 days
	var mockActivities []domain.UserActivityTracking
	for i := 0; i < 30; i++ {
		d := dateNow.AddDate(0, 0, -i).Truncate(24 * time.Hour)
		
		for _, m := range allModules {
			// Random view count between 5 and 50
			views := 5 + ((i + len(m.NameEn)) % 45)
			
			moduleUUID, _ := uuid.Parse(m.ID)
			mockActivities = append(mockActivities, domain.UserActivityTracking{
				ID:             uuid.New(),
				Date:           d,
				ModuleId:       moduleUUID,
				ViewCount:      views,
			})
		}
	}
	r.db.Create(&mockActivities)

	return nil
}
