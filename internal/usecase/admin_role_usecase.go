package usecase

import (
	"super-app-chonburi-go/internal/domain"
)

type adminRoleUseCase struct {
	repo domain.AdminRoleRepository
}

func NewAdminRoleUseCase(repo domain.AdminRoleRepository) domain.AdminRoleUseCase {
	return &adminRoleUseCase{repo: repo}
}

func (u *adminRoleUseCase) GetAllRoles() ([]domain.AdminRole, error) {
	return u.repo.GetAll()
}
