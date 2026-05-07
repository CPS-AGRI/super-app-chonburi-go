package usecase

import (
	"super-app-chonburi-go/internal/domain"
)

type dashboardUseCase struct {
	repo domain.DashboardRepository
}

func NewDashboardUseCase(repo domain.DashboardRepository) domain.DashboardUseCase {
	return &dashboardUseCase{repo: repo}
}

func (u *dashboardUseCase) GetBackOffice(filter domain.DashboardFilter) (*domain.DashboardBackOfficeModuleDTO, error) {
	return u.repo.GetBackOffice(filter)
}

func (u *dashboardUseCase) SeedMockData(municipalityId string) error {
	return u.repo.SeedMockData(municipalityId)
}
