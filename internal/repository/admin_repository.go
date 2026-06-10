package repository

import (
	"super-app-chonburi-go/internal/domain"

	"gorm.io/gorm"
)

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) domain.AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetByEmail(email string) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.db.Preload("Role").
		Preload("Departments").
		Preload("Departments.Modules").
		Preload("Departments.Modules.ModuleTypes").
		Where("email = ?", email).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetByID(id string) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.db.Preload("Role").
		Preload("Departments").
		Preload("Departments.Modules").
		Preload("Departments.Modules.ModuleTypes").
		Where("id = ?", id).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetPaginated(query domain.AdminQuery) (*domain.PaginatedAdminResponse, error) {
	var admins []domain.Admin
	var total int64

	db := r.db.Model(&domain.Admin{})
	if query.Name != "" {
		db = db.Where("name ILIKE ?", "%"+query.Name+"%")
	}
	if query.Email != "" {
		db = db.Where("email ILIKE ?", "%"+query.Email+"%")
	}

	db.Count(&total)
	offset := (query.PageNumber - 1) * query.PageSize
	err := r.db.Model(&domain.Admin{}).
		Preload("Role").
		Preload("Departments").
		Offset(offset).
		Limit(query.PageSize).
		Find(&admins).Error

	return &domain.PaginatedAdminResponse{
		Items:      admins,
		TotalItems: total,
		PageNumber: query.PageNumber,
		TotalPages: int((total + int64(query.PageSize) - 1) / int64(query.PageSize)),
	}, err
}

func (r *adminRepository) Create(admin *domain.Admin) error {

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(admin).Error; err != nil {
			return err
		}

		return tx.Model(admin).Association("Departments").Replace(admin.Departments)
	})
}

func (r *adminRepository) Update(admin *domain.Admin) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(admin).Association("Departments").Replace(admin.Departments); err != nil {
			return err
		}

		return tx.Save(admin).Error
	})
}

func (r *adminRepository) Delete(id string) error {
	return r.db.Delete(&domain.Admin{}, "id = ?", id).Error
}
