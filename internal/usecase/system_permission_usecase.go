package usecase

import (
	"super-app-chonburi-go/internal/domain"
)

type systemPermissionUseCase struct {
	repo domain.SystemPermissionRepository
}

func NewSystemPermissionUseCase(repo domain.SystemPermissionRepository) domain.SystemPermissionUseCase {
	return &systemPermissionUseCase{repo: repo}
}

func (u *systemPermissionUseCase) GetAllPermissions() ([]domain.SystemPermission, error) {
	return u.repo.GetAll()
}
