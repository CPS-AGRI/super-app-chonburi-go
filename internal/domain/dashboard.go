package domain

import "time"

type PersentUsersDTO struct {
	PersentMaleUsers   float64 `json:"persent_male_users"`
	PersentFemaleUsers float64 `json:"persent_female_users"`
}

type AgeGroupDTO struct {
	AgeGroup20To30     int `json:"age_group_20_to_30"`
	AgeGroup31To40     int `json:"age_group_31_to_40"`
	AgeGroup41To50     int `json:"age_group_41_to_50"`
	AgeGroup51To60     int `json:"age_group_51_to_60"`
	AgeGroup61AndAbove int `json:"age_group_61_and_above"`
}

type AgeGenderDTO struct {
	Male   AgeGroupDTO `json:"male"`
	Female AgeGroupDTO `json:"female"`
}

type DailyUserDTO struct {
	Date                             time.Time `json:"date"`
	TotalDailyUserUsageCount         int       `json:"total_daily_user_usage_count"`
	TotalDailyVerifiedUserUsageCount int       `json:"total_daily_verified_user_usage_count"`
	TotalDailyLoginUserUsageCount    int       `json:"total_daily_login_user_usage_count"`
}

type ModuleNameUsageDTO struct {
	Th string `json:"th"`
	En string `json:"en"`
}

type ModuleUsageDTO struct {
	ModuleId          string             `json:"module_id"`
	ModuleDisplayName ModuleNameUsageDTO `json:"module_display_name"`
	ModuleUsageCount  int                `json:"module_usage_count"`
}

type DashboardBackOfficeModuleDTO struct {
	TotalUsers           int              `json:"total_users"`
	TotalVerifiedUsers   int              `json:"total_verified_users"`
	TotalLoginUsersCount int              `json:"total_login_users_count"`
	PersentUsers         PersentUsersDTO  `json:"persent_users"`
	AgeGender            AgeGenderDTO     `json:"age_gender"`
	TotalModuleUsage     int              `json:"total_module_usage"`
	DailyUsers           []DailyUserDTO   `json:"daily_users"`
	ModuleUsage          []ModuleUsageDTO `json:"module_usage"`
}

type DashboardFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
}

type DashboardRepository interface {
	GetBackOffice(filter DashboardFilter) (*DashboardBackOfficeModuleDTO, error)
	SeedMockData(municipalityId string) error
}

type DashboardUseCase interface {
	GetBackOffice(filter DashboardFilter) (*DashboardBackOfficeModuleDTO, error)
	SeedMockData(municipalityId string) error
}
