package repository

import (
	"errors"
	"fmt"
	"super-app-chonburi-go/internal/domain"
	"time"

	"gorm.io/gorm"
)

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) domain.DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) GetPaginated(query domain.DepartmentQuery) (*domain.PaginatedDepartmentResponse, error) {
	var depts []domain.Department
	var total int64

	db := r.db.Model(&domain.Department{})
	if query.Name != "" {
		db = db.Where("name ILIKE ?", "%"+query.Name+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	db.Count(&total)
	offset := (query.PageNumber - 1) * query.PageSize
	err := db.Offset(offset).Limit(query.PageSize).Order("created_date DESC").Find(&depts).Error
	if err == nil {
		for i := range depts {
			r.populateModules(r.db, &depts[i])
		}
	}

	return &domain.PaginatedDepartmentResponse{
		Items:      depts,
		TotalItems: total,
		PageNumber: query.PageNumber,
		TotalPages: int((total + int64(query.PageSize) - 1) / int64(query.PageSize)),
	}, err
}

func (r *departmentRepository) GetByID(id string) (*domain.Department, error) {
	var dept domain.Department
	if err := r.db.First(&dept, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if err := r.populateModules(r.db, &dept); err != nil {
		return nil, err
	}

	return &dept, nil
}

func (r *departmentRepository) GetAll() ([]domain.Department, error) {
	var depts []domain.Department
	if err := r.db.Order("name ASC").Find(&depts).Error; err != nil {
		return nil, err
	}

	for i := range depts {
		if err := r.populateModules(r.db, &depts[i]); err != nil {
			return nil, err
		}
	}

	return depts, nil
}

func (r *departmentRepository) populateModules(tx *gorm.DB, dept *domain.Department) error {
	// Find all DepartmentModules for this department
	var deptModules []domain.DepartmentModule
	if err := tx.Where("department_id = ?", dept.ID).Find(&deptModules).Error; err != nil {
		return err
	}

	if len(deptModules) == 0 {
		dept.Modules = []domain.Module{}
		return nil
	}

	dept.Modules = make([]domain.Module, 0, len(deptModules))
	for _, dm := range deptModules {
		var module domain.Module
		if err := tx.First(&module, "id = ?", dm.ModuleId).Error; err != nil {
			continue
		}

		// Find assigned ModuleTypes for this DepartmentModule
		var typeIDs []string
		tx.Model(&domain.DepartmentModuleModuleType{}).
			Where("department_module_id = ?", dm.ID).
			Pluck("module_type_id", &typeIDs)

		if len(typeIDs) > 0 {
			var moduleTypes []domain.ModuleType
			tx.Where("id IN ?", typeIDs).Find(&moduleTypes)
			module.ModuleTypes = moduleTypes
		}

		dept.Modules = append(dept.Modules, module)
	}

	return nil
}

func (r *departmentRepository) Create(dept *domain.Department) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(dept).Error; err != nil {
			return err
		}
		return r.assignModuleTypes(tx, dept)
	})
}

func (r *departmentRepository) Update(dept *domain.Department) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(dept).Error; err != nil {
			return err
		}
		// Clear existing bridge data
		var deptModules []domain.DepartmentModule
		tx.Where("department_id = ?", dept.ID).Find(&deptModules)
		for _, dm := range deptModules {
			tx.Where("department_module_id = ?", dm.ID).Delete(&domain.DepartmentModuleModuleType{})
		}
		tx.Where("department_id = ?", dept.ID).Delete(&domain.DepartmentModule{})

		return r.assignModuleTypes(tx, dept)
	})
}

func (r *departmentRepository) assignModuleTypes(tx *gorm.DB, dept *domain.Department) error {
	fmt.Printf("DEBUG: assignModuleTypes for Dept: %s\n", dept.ID)
	fmt.Printf("DEBUG: ModuleIds: %v, ModuleTypeIds: %v\n", dept.ModuleIds, dept.ModuleTypeIds)

	// 1. Handle Module Assignments (Direct)
	uniqueModuleIDs := make(map[string]bool)
	for _, id := range dept.ModuleIds {
		if id != "" {
			uniqueModuleIDs[id] = true
		}
	}

	// Also include modules from module types
	if len(dept.ModuleTypeIds) > 0 {
		var moduleTypes []domain.ModuleType
		if err := tx.Where("id IN ?", dept.ModuleTypeIds).Find(&moduleTypes).Error; err != nil {
			fmt.Printf("ERROR: Failed to find module types: %v\n", err)
			return err
		}
		for _, mt := range moduleTypes {
			uniqueModuleIDs[mt.ModuleId] = true
		}
	}

	fmt.Printf("DEBUG: Unique Module IDs to assign: %v\n", uniqueModuleIDs)

	// Create DepartmentModule records
	moduleToDMID := make(map[string]string)
	for mID := range uniqueModuleIDs {
		dm := domain.DepartmentModule{
			ID:           domain.NewUUID(),
			DepartmentId: dept.ID,
			ModuleId:     mID,
			CreatedBy:    dept.CreatedBy,
			UpdatedBy:    dept.UpdatedBy,
			CreatedDate:  time.Now(),
			UpdatedDate:  time.Now(),
		}
		if err := tx.Create(&dm).Error; err != nil {
			fmt.Printf("ERROR: Failed to create department_module: %v\n", err)
			return err
		}
		moduleToDMID[mID] = dm.ID
		fmt.Printf("DEBUG: Created department_module %s for module %s\n", dm.ID, mID)
	}

	// 2. Handle ModuleType Assignments
	uniqueTypeIDs := make(map[string]bool)
	for _, id := range dept.ModuleTypeIds {
		if id != "" {
			uniqueTypeIDs[id] = true
		}
	}

	// AUTOMATICALLY add mandatory types for all assigned modules
	var allModuleIDs []string
	for id := range uniqueModuleIDs {
		allModuleIDs = append(allModuleIDs, id)
	}
	if len(allModuleIDs) > 0 {
		var mandatoryTypes []domain.ModuleType
		tx.Where("module_id IN ? AND can_be_selected_with_admin_user_settings = ?", allModuleIDs, false).Find(&mandatoryTypes)
		for _, mt := range mandatoryTypes {
			uniqueTypeIDs[mt.ID] = true
			fmt.Printf("DEBUG: Auto-adding mandatory type %s for module %s\n", mt.ID, mt.ModuleId)
		}
	}

	if len(uniqueTypeIDs) > 0 {
		var finalTypeIDs []string
		for id := range uniqueTypeIDs {
			finalTypeIDs = append(finalTypeIDs, id)
		}

		var moduleTypes []domain.ModuleType
		tx.Where("id IN ?", finalTypeIDs).Find(&moduleTypes)
		
		for _, mt := range moduleTypes {
			dmID, ok := moduleToDMID[mt.ModuleId]
			if !ok {
				continue
			}
			dmmt := domain.DepartmentModuleModuleType{
				DepartmentModuleId: dmID,
				ModuleTypeId:       mt.ID,
			}
			if err := tx.Create(&dmmt).Error; err != nil {
				fmt.Printf("ERROR: Failed to create department_module_module_type: %v\n", err)
				return err
			}
			fmt.Printf("DEBUG: Linked ModuleType %s to DeptModule %s\n", mt.ID, dmID)
		}
	}

	return nil
}

func (r *departmentRepository) Delete(id string) error {
	return r.db.Delete(&domain.Department{}, "id = ?", id).Error
}
