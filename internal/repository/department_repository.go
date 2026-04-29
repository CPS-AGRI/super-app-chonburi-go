package repository

import (
	"errors"
	"math"

	"github.com/google/uuid"
	"super-app-chonburi-go/internal/domain"
	"gorm.io/gorm"
)

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) domain.DepartmentRepository {
	return &departmentRepository{db}
}

func (r *departmentRepository) GetPaginated(query domain.DepartmentQuery) (*domain.PaginatedDepartmentResponse, error) {
	var depts []domain.Department
	var totalItems int64

	db := r.db.Model(&domain.Department{})

	if query.Name != "" {
		db = db.Where("name ILIKE ?", "%"+query.Name+"%")
	}

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	db.Count(&totalItems)

	offset := (query.PageNumber - 1) * query.PageSize
	err := db.Preload("Permissions").Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&depts).Error
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(query.PageSize)))

	return &domain.PaginatedDepartmentResponse{
		PageNumber: query.PageNumber,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Items:      depts,
	}, nil
}

func (r *departmentRepository) GetByID(id uuid.UUID) (*domain.Department, error) {
	var dept domain.Department
	err := r.db.Preload("Permissions").First(&dept, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &dept, nil
}

func (r *departmentRepository) GetAll() ([]domain.Department, error) {
	var depts []domain.Department
	err := r.db.Preload("Permissions").Order("name ASC").Find(&depts).Error
	return depts, err
}

func (r *departmentRepository) Create(dept *domain.Department) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(dept).Error; err != nil {
			return err
		}

		if len(dept.PermissionIDs) > 0 {
			var perms []domain.SystemPermission
			if err := tx.Where("id IN ?", dept.PermissionIDs).Find(&perms).Error; err != nil {
				return err
			}
			if err := tx.Model(dept).Association("Permissions").Replace(perms); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *departmentRepository) Update(dept *domain.Department) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(dept).Omit("CreatedAt", "CreatedBy").Updates(dept).Error; err != nil {
			return err
		}

		// Update permissions
		var perms []domain.SystemPermission
		if len(dept.PermissionIDs) > 0 {
			if err := tx.Where("id IN ?", dept.PermissionIDs).Find(&perms).Error; err != nil {
				return err
			}
		}
		return tx.Model(dept).Association("Permissions").Replace(perms)
	})
}

func (r *departmentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Department{}, "id = ?", id).Error
}
