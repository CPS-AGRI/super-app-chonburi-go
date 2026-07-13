package usecase

import (
	"super-app-chonburi-go/internal/domain"
	"time"
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

func (u *adminRoleUseCase) GetRoleByID(id string) (*domain.AdminRole, error) {
	return u.repo.GetByID(id)
}
func (u *adminRoleUseCase) CreateRole(role *domain.AdminRole) error {
	role.ID = domain.NewUUID()
	role.CreatedDate = time.Now()
	role.UpdatedDate = time.Now()
	return u.repo.Create(role)
}

func (u *adminRoleUseCase) UpdateRole(role *domain.AdminRole) error {
	role.UpdatedDate = time.Now()
	return u.repo.Update(role)
}

func (u *adminRoleUseCase) DeleteRole(id string) error {
	return u.repo.Delete(id)
}
