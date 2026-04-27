package repository

import (
	"errors"
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
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
	if err := r.db.Preload("Department").First(&admin, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetByID(id uuid.UUID) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.Preload("Department").First(&admin, "id = ?", id).Error; err != nil {
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

	dbQuery := r.db.Model(&domain.Admin{}).
		Where("\"departmentId\" NOT IN (SELECT id FROM \"AdminDepartment\" WHERE name = ?)", "superadmin")

	if query.Email != "" {
		dbQuery = dbQuery.Where("email ILIKE ?", "%"+query.Email+"%")
	}
	if query.Name != "" {
		dbQuery = dbQuery.Where("name ILIKE ?", "%"+query.Name+"%")
	}
	if query.DepartmentID != "" {
		dbQuery = dbQuery.Where("\"departmentId\" = ?", query.DepartmentID)
	}

	if err := dbQuery.Count(&totalItems).Error; err != nil {
		return nil, err
	}

	totalPages := int((totalItems + int64(query.PageSize) - 1) / int64(query.PageSize))
	offset := (query.PageNumber - 1) * query.PageSize

	if err := dbQuery.Preload("Department").Offset(offset).Limit(query.PageSize).Order("\"createdAt\" DESC").Find(&admins).Error; err != nil {
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
	return r.db.Create(admin).Error
}

func (r *adminRepository) Update(admin *domain.Admin) error {
	return r.db.Save(admin).Error
}

func (r *adminRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Admin{}, "id = ?", id).Error
}
