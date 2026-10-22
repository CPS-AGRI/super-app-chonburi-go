package usecase

import (
	"strings"
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

	if admin.Role != nil && (strings.EqualFold(admin.Role.Type, domain.RoleSuperAdmin) || strings.EqualFold(admin.Role.Type, "super_admin")) {
		return u.repo.GetAll()
	}

	isExecutive := admin.Role != nil && (strings.EqualFold(admin.Role.Type, domain.RoleExecutives) || strings.EqualFold(admin.Role.Type, "executive"))

	// Collect unique modules from assigned departments
	moduleMap := make(map[string]domain.Module)
	for _, dept := range admin.Departments {
		for _, mod := range dept.Modules {
			// For Executive role, only return modules that have an active dashboard in DB
			if isExecutive && !mod.IsDashboard {
				continue
			}
			moduleMap[mod.ID] = mod
		}
	}

	modules := make([]domain.Module, 0, len(moduleMap))
	for _, mod := range moduleMap {
		modules = append(modules, mod)
	}

	return modules, nil
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
