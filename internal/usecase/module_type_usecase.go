package usecase

import (
	"super-app-chonburi-go/internal/domain"
	"time"
)

type moduleTypeUseCase struct {
	moduleTypeRepo domain.ModuleTypeRepository
}

func NewModuleTypeUseCase(repo domain.ModuleTypeRepository) domain.ModuleTypeUseCase {
	return &moduleTypeUseCase{moduleTypeRepo: repo}
}

func (u *moduleTypeUseCase) GetAllTypes() ([]domain.ModuleType, error) {
	return u.moduleTypeRepo.GetAll()
}

func (u *moduleTypeUseCase) GetTypesByModule(moduleID string) ([]domain.ModuleType, error) {
	return u.moduleTypeRepo.GetByModuleID(moduleID)
}

func (u *moduleTypeUseCase) AssignTypesToDepartmentModule(deptID, moduleID string, typeIDs []string) error {
	return u.moduleTypeRepo.AssignToDepartmentModule(deptID, moduleID, typeIDs)
}

func (u *moduleTypeUseCase) CreateType(mt *domain.ModuleType) error {
	mt.ID = domain.NewUUID()
	mt.CreatedDate = time.Now()
	mt.UpdatedDate = time.Now()
	mt.CreatedBy = "system"
	mt.UpdatedBy = "system"
	return u.moduleTypeRepo.Create(mt)
}

func (u *moduleTypeUseCase) UpdateType(mt *domain.ModuleType) error {
	mt.UpdatedDate = time.Now()
	return u.moduleTypeRepo.Update(mt)
}

func (u *moduleTypeUseCase) DeleteType(id string) error {
	return u.moduleTypeRepo.Delete(id)
}
