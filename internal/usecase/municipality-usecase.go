package usecase

import (
	"errors"
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
)

type municipalityUseCase struct {
	repo domain.MunicipalityRepository
}

func NewMunicipalityUseCase(repo domain.MunicipalityRepository) domain.MunicipalityUseCase {
	return &municipalityUseCase{repo: repo}
}

func (u *municipalityUseCase) GetList() ([]domain.Municipality, error) {
	return u.repo.GetList()
}

func (u *municipalityUseCase) GetDetail(id uuid.UUID) (*domain.Municipality, error) {
	return u.repo.GetByID(id)
}

func (u *municipalityUseCase) Create(muni *domain.Municipality) error {
	muni.ID = uuid.New()
	return u.repo.Create(muni)
}

func (u *municipalityUseCase) Update(muni *domain.Municipality) error {
	return u.repo.Update(muni)
}

func (u *municipalityUseCase) Delete(id uuid.UUID) error {
	return u.repo.Delete(id)
}

func (u *municipalityUseCase) GetCurrent() (*domain.Municipality, error) {
	return u.repo.GetFirst()
}

// Error definitions
var (
	ErrMunicipalityNotFound = errors.New("municipality not found")
)
