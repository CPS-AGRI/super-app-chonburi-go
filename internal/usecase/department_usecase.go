package usecase

import (
	"github.com/google/uuid"
	"super-app-chonburi-go/internal/domain"
)

type departmentUseCase struct {
	deptRepo domain.DepartmentRepository
}

func NewDepartmentUseCase(deptRepo domain.DepartmentRepository) domain.DepartmentUseCase {
	return &departmentUseCase{deptRepo}
}

func (u *departmentUseCase) GetDepartments(query domain.DepartmentQuery) (*domain.PaginatedDepartmentResponse, error) {
	if query.PageNumber <= 0 {
		query.PageNumber = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	return u.deptRepo.GetPaginated(query)
}

func (u *departmentUseCase) GetAllDepartments() ([]domain.Department, error) {
	return u.deptRepo.GetAll()
}

func (u *departmentUseCase) GetDepartmentByID(id uuid.UUID) (*domain.Department, error) {
	return u.deptRepo.GetByID(id)
}

func (u *departmentUseCase) CreateDepartment(dept *domain.Department) error {
	return u.deptRepo.Create(dept)
}

func (u *departmentUseCase) UpdateDepartment(dept *domain.Department) error {
	return u.deptRepo.Update(dept)
}

func (u *departmentUseCase) DeleteDepartment(id uuid.UUID) error {
	return u.deptRepo.Delete(id)
}
