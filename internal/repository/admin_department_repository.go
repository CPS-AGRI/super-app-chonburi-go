package repository

import (
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type adminDepartmentRepository struct {
	db *gorm.DB
}

func NewAdminDepartmentRepository(db *gorm.DB) domain.AdminDepartmentRepository {
	return &adminDepartmentRepository{db: db}
}

func (r *adminDepartmentRepository) GetPaginated(query domain.AdminDepartmentQuery) (*domain.PaginatedAdminDepartmentResponse, error) {
	var departments []domain.AdminDepartment
	var totalItems int64

	dbQuery := r.db.Model(&domain.AdminDepartment{})

	if query.Name != "" {
		dbQuery = dbQuery.Where("name ILIKE ?", "%"+query.Name+"%")
	}

	if err := dbQuery.Count(&totalItems).Error; err != nil {
		return nil, err
	}

	totalPages := int((totalItems + int64(query.PageSize) - 1) / int64(query.PageSize))
	offset := (query.PageNumber - 1) * query.PageSize

	if err := dbQuery.Offset(offset).Limit(query.PageSize).Order("\"createdAt\" DESC").Find(&departments).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedAdminDepartmentResponse{
		PageNumber: query.PageNumber,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Items:      departments,
	}, nil
}

func (r *adminDepartmentRepository) GetByID(id uuid.UUID) (*domain.AdminDepartment, error) {
	var department domain.AdminDepartment
	err := r.db.First(&department, "id = ?", id).Error
	return &department, err
}

func (r *adminDepartmentRepository) Create(department *domain.AdminDepartment) error {
	return r.db.Create(department).Error
}

func (r *adminDepartmentRepository) Update(department *domain.AdminDepartment) error {
	return r.db.Save(department).Error
}

func (r *adminDepartmentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.AdminDepartment{}, "id = ?", id).Error
}
