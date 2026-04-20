package repository

import (
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type municipalityRepository struct {
	db *gorm.DB
}

func NewMunicipalityRepository(db *gorm.DB) domain.MunicipalityRepository {
	return &municipalityRepository{db: db}
}

func (r *municipalityRepository) GetList() ([]domain.Municipality, error) {
	var list []domain.Municipality
	err := r.db.Order("createdAt desc").Find(&list).Error
	return list, err
}

func (r *municipalityRepository) GetByID(id uuid.UUID) (*domain.Municipality, error) {
	var muni domain.Municipality
	err := r.db.First(&muni, "id = ?", id).Error
	return &muni, err
}

func (r *municipalityRepository) Create(muni *domain.Municipality) error {
	return r.db.Create(muni).Error
}

func (r *municipalityRepository) Update(muni *domain.Municipality) error {
	return r.db.Save(muni).Error
}

func (r *municipalityRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Municipality{}, "id = ?", id).Error
}

func (r *municipalityRepository) GetFirst() (*domain.Municipality, error) {
	var muni domain.Municipality
	err := r.db.First(&muni).Error
	if err != nil {
		return nil, err
	}
	return &muni, nil
}
