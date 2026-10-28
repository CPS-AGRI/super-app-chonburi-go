package usecase

import (
	"super-app-chonburi-go/internal/domain"
)

type kpiUseCase struct {
	repo domain.KPILogRepository
}

func NewKPIUseCase(repo domain.KPILogRepository) domain.KPIUseCase {
	return &kpiUseCase{repo: repo}
}

func (u *kpiUseCase) GetSummary(filter domain.KPIQueryFilter) (*domain.KPISummaryResponse, error) {
	return u.repo.GetKPISummary(filter)
}

func (u *kpiUseCase) GetStaffLeaderboard(filter domain.KPIQueryFilter) ([]domain.KPIStaffLeaderboardItem, error) {
	return u.repo.GetStaffLeaderboard(filter)
}

func (u *kpiUseCase) GetManagerScorecard(filter domain.KPIQueryFilter) ([]domain.KPIManagerScorecardItem, error) {
	return u.repo.GetManagerScorecard(filter)
}

func (u *kpiUseCase) GetStaffDetail(userID string, filter domain.KPIQueryFilter) (*domain.KPIStaffDetailResponse, error) {
	return u.repo.GetStaffDetail(userID, filter)
}

func (u *kpiUseCase) GetManagerDetail(userID string, filter domain.KPIQueryFilter) (*domain.KPIManagerDetailResponse, error) {
	return u.repo.GetManagerDetail(userID, filter)
}

func (u *kpiUseCase) GetBreachedCases(filter domain.KPIQueryFilter) ([]domain.KPIBreachedCaseItem, error) {
	return u.repo.GetBreachedCases(filter)
}

func (u *kpiUseCase) GetCategoryStats(filter domain.KPIQueryFilter) ([]domain.KPICategoryStatItem, error) {
	return u.repo.GetCategoryStats(filter)
}

func (u *kpiUseCase) GetThresholds(moduleKey string) (*domain.KPIThresholdConfig, error) {
	return u.repo.GetThresholds(moduleKey)
}

func (u *kpiUseCase) UpdateThresholds(moduleKey string, req domain.UpdateKPIThresholdRequest, adminID string) (*domain.KPIThresholdConfig, error) {
	return u.repo.UpdateThresholds(moduleKey, req, adminID)
}

func (u *kpiUseCase) GetDataSharingLogs(module string, startDate, endDate string, limit int) ([]domain.DataSharingLog, error) {
	return u.repo.GetDataSharingLogs(module, startDate, endDate, limit)
}

func (u *kpiUseCase) GetDataSharingSummary(module string, startDate, endDate string) (*domain.DataSharingSummaryResponse, error) {
	return u.repo.GetDataSharingSummary(module, startDate, endDate)
}
