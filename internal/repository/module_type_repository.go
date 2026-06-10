package repository

import (
	"super-app-chonburi-go/internal/domain"

	"gorm.io/gorm"
)

type moduleTypeRepository struct {
	db *gorm.DB
}

func NewModuleTypeRepository(db *gorm.DB) domain.ModuleTypeRepository {
	return &moduleTypeRepository{db: db}
}

func (r *moduleTypeRepository) GetAll() ([]domain.ModuleType, error) {
	var types []domain.ModuleType
	err := r.db.Find(&types).Error
	return types, err
}

func (r *moduleTypeRepository) GetByModuleID(moduleID string) ([]domain.ModuleType, error) {
	var types []domain.ModuleType
	err := r.db.Where("module_id = ?", moduleID).Find(&types).Error
	return types, err
}

func (r *moduleTypeRepository) GetByDepartmentModule(deptID, moduleID string) ([]domain.ModuleType, error) {
	var types []domain.ModuleType
	err := r.db.Joins("JOIN department_module_module_types ON department_module_module_types.module_type_id = module_types.id").
		Joins("JOIN department_modules ON department_modules.id = department_module_module_types.department_module_id").
		Where("department_modules.department_id = ? AND department_modules.module_id = ?", deptID, moduleID).
		Find(&types).Error
	return types, err
}

func (r *moduleTypeRepository) AssignToDepartmentModule(deptID, moduleID string, typeIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		var deptModule domain.DepartmentModule
		if err := tx.Where("department_id = ? AND module_id = ?", deptID, moduleID).First(&deptModule).Error; err != nil {
			return err
		}

		if err := tx.Where("department_module_id = ?", deptModule.ID).
			Delete(&domain.DepartmentModuleModuleType{}).Error; err != nil {
			return err
		}

		for _, tID := range typeIDs {
			assignment := domain.DepartmentModuleModuleType{
				DepartmentModuleId: deptModule.ID,
				ModuleTypeId:       tID,
			}
			if err := tx.Create(&assignment).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *moduleTypeRepository) Create(mt *domain.ModuleType) error {
	return r.db.Create(mt).Error
}

func (r *moduleTypeRepository) Update(mt *domain.ModuleType) error {
	return r.db.Model(mt).Updates(mt).Error
}

func (r *moduleTypeRepository) Delete(id string) error {
	return r.db.Delete(&domain.ModuleType{}, "id = ?", id).Error
}
