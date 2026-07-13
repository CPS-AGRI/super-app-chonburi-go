package usecase

import (
	"super-app-chonburi-go/internal/domain"
	"time"
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

func (u *departmentUseCase) GetDepartmentByID(id string) (*domain.Department, error) {
	return u.deptRepo.GetByID(id)
}

func (u *departmentUseCase) CreateDepartment(dept *domain.Department) error {
	dept.ID = domain.NewUUID()
	dept.CreatedDate = time.Now()
	dept.UpdatedDate = time.Now()
	return u.deptRepo.Create(dept)
}

func (u *departmentUseCase) UpdateDepartment(dept *domain.Department) error {
	dept.UpdatedDate = time.Now()
	return u.deptRepo.Update(dept)
}

func (u *departmentUseCase) DeleteDepartment(id string) error {
	return u.deptRepo.Delete(id)
}
