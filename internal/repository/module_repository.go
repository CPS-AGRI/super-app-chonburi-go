package repository

import (
	"super-app-chonburi-go/internal/domain"

	"gorm.io/gorm"
)

type moduleRepository struct {
	db *gorm.DB
}

func NewModuleRepository(db *gorm.DB) domain.ModuleRepository {
	return &moduleRepository{db: db}
}

func (r *moduleRepository) GetAll() ([]domain.Module, error) {
	var modules []domain.Module
	err := r.db.Preload("ModuleTypes").Order("sequence ASC").Find(&modules).Error
	return modules, err
}

func (r *moduleRepository) GetByDepartmentID(deptID string) ([]domain.Module, error) {
	var modules []domain.Module
	err := r.db.Preload("ModuleTypes").
		Joins("JOIN department_modules ON department_modules.module_id = modules.id").
		Where("department_modules.department_id = ?", deptID).
		Find(&modules).Error
	return modules, err
}

func (r *moduleRepository) AssignToDepartment(deptID string, moduleIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Where("department_id = ?", deptID).Delete(&domain.DepartmentModule{}).Error; err != nil {
			return err
		}

		for _, mID := range moduleIDs {
			assignment := domain.DepartmentModule{
				DepartmentId: deptID,
				ModuleId:     mID,
			}
			if err := tx.Create(&assignment).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *moduleRepository) Create(module *domain.Module) error {
	return r.db.Create(module).Error
}

func (r *moduleRepository) Update(module *domain.Module) error {
	return r.db.Model(module).Updates(module).Error
}

func (r *moduleRepository) Delete(id string) error {
	return r.db.Delete(&domain.Module{}, "id = ?", id).Error
}
