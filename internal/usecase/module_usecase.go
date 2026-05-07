package usecase

import (
	"super-app-chonburi-go/internal/domain"
	"time"
)

type moduleUseCase struct {
	repo      domain.ModuleRepository
	adminRepo domain.AdminRepository
}

func NewModuleUseCase(repo domain.ModuleRepository, adminRepo domain.AdminRepository) domain.ModuleUseCase {
	return &moduleUseCase{
		repo:      repo,
		adminRepo: adminRepo,
	}
}

func (u *moduleUseCase) GetModulesForUser(adminID string) ([]domain.Module, error) {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	// If SuperAdmin, return all modules
	if admin.Role != nil && admin.Role.Type == "super_admin" {
		return u.repo.GetAll()
	}

	return []domain.Module{}, nil
}

func (u *moduleUseCase) GetAllModules() ([]domain.Module, error) {
	return u.repo.GetAll()
}

func (u *moduleUseCase) AssignModulesToDepartment(deptID string, moduleIDs []string) error {
	return u.repo.AssignToDepartment(deptID, moduleIDs)
}
func (u *moduleUseCase) CreateModule(module *domain.Module) error {
	module.ID = domain.NewUUID()
	module.CreatedBy = "system"
	module.UpdatedBy = "system"
	module.CreatedDate = time.Now()
	module.UpdatedDate = time.Now()
	return u.repo.Create(module)
}

func (u *moduleUseCase) UpdateModule(module *domain.Module) error {
	return u.repo.Update(module)
}

func (u *moduleUseCase) DeleteModule(id string) error {
	return u.repo.Delete(id)
}
