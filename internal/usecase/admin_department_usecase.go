package usecase

import (
	"errors"
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
)

type adminDepartmentUseCase struct {
	repo domain.AdminDepartmentRepository
}

func NewAdminDepartmentUseCase(repo domain.AdminDepartmentRepository) domain.AdminDepartmentUseCase {
	return &adminDepartmentUseCase{repo: repo}
}

func (u *adminDepartmentUseCase) GetDepartments(query domain.AdminDepartmentQuery) (*domain.PaginatedAdminDepartmentResponse, error) {
	if query.PageNumber < 1 {
		query.PageNumber = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	return u.repo.GetPaginated(query)
}

func (u *adminDepartmentUseCase) GetDepartmentByID(id uuid.UUID) (*domain.AdminDepartment, error) {
	return u.repo.GetByID(id)
}

func (u *adminDepartmentUseCase) CreateDepartment(department *domain.AdminDepartment) error {
	if department.Name == "" {
		return errors.New("department name is required")
	}
	return u.repo.Create(department)
}

func (u *adminDepartmentUseCase) UpdateDepartment(department *domain.AdminDepartment) error {
	existing, err := u.repo.GetByID(department.ID)
	if err != nil {
		return errors.New("department not found")
	}

	if department.Name == "" {
		return errors.New("department name cannot be empty")
	}

	existing.Name = department.Name
	existing.Description = department.Description
	existing.IsActive = department.IsActive
	existing.Permissions = department.Permissions

	return u.repo.Update(existing)
}

func (u *adminDepartmentUseCase) DeleteDepartment(id uuid.UUID) error {
	_, err := u.repo.GetByID(id)
	if err != nil {
		return errors.New("department not found")
	}
	// In a real scenario, check if admins exist before deleting to avoid violating references
	return u.repo.Delete(id)
}
