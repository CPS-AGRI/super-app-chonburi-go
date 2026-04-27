package repository

import (
	"errors"
	"math"

	"github.com/google/uuid"
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
	err := r.db.Preload("Role").Preload("Departments").Preload("Departments.Permissions").First(&admin, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetByID(id uuid.UUID) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.db.Preload("Role").Preload("Departments").First(&admin, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetPaginated(query domain.AdminQuery) (*domain.PaginatedAdminResponse, error) {
	var admins []domain.Admin
	var totalItems int64

	db := r.db.Model(&domain.Admin{}).
		Joins("JOIN admin_roles ON admins.role_id = admin_roles.id").
		Where("admin_roles.is_superadmin = ?", false)

	if query.Email != "" {
		db = db.Where("admins.email ILIKE ?", "%"+query.Email+"%")
	}
	if query.Name != "" {
		db = db.Where("admins.name ILIKE ?", "%"+query.Name+"%")
	}

	db.Count(&totalItems)

	totalPages := int(math.Ceil(float64(totalItems) / float64(query.PageSize)))
	offset := (query.PageNumber - 1) * query.PageSize

	err := db.Preload("Role").Preload("Departments").Offset(offset).Limit(query.PageSize).Find(&admins).Error
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedAdminResponse{
		PageNumber: query.PageNumber,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Items:      admins,
	}, nil
}

func (r *adminRepository) Create(admin *domain.Admin) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(admin).Error; err != nil {
			return err
		}

		// Handle Departments Many-to-Many
		if len(admin.Departments) > 0 {
			if err := tx.Model(admin).Association("Departments").Replace(admin.Departments); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *adminRepository) Update(admin *domain.Admin) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(admin).Error; err != nil {
			return err
		}

		// Sync Departments
		return tx.Model(admin).Association("Departments").Replace(admin.Departments)
	})
}

func (r *adminRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Admin{}, "id = ?", id).Error
}
