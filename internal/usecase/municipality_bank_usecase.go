package usecase

import (
	"super-app-chonburi-go/internal/domain"
	"time"

	"github.com/google/uuid"
)

type municipalityBankUseCase struct {
	repo domain.MunicipalityBankRepository
}

func NewMunicipalityBankUseCase(repo domain.MunicipalityBankRepository) domain.MunicipalityBankUseCase {
	return &municipalityBankUseCase{repo: repo}
}

func (u *municipalityBankUseCase) GetAllBanks() ([]domain.MunicipalityBank, error) {
	return u.repo.GetAll()
}

func (u *municipalityBankUseCase) GetActiveBank() (*domain.MunicipalityBank, error) {
	return u.repo.GetActive()
}

func (u *municipalityBankUseCase) SaveBank(bank *domain.MunicipalityBank) error {
	if bank.ID == uuid.Nil {
		bank.ID = uuid.New()
		bank.CreatedAt = time.Now()
		bank.UpdatedAt = time.Now()
		return u.repo.Create(bank)
	}

	existing, err := u.repo.GetByID(bank.ID)
	if err != nil {
		return err
	}

	bank.CreatedAt = existing.CreatedAt
	bank.UpdatedAt = time.Now()
	
	// Ensure we don't accidentally wipe prompt_pay_number if not provided in update, or allow it
	if bank.PromptPayNumber == "" && existing.PromptPayNumber != "" {
		// Depending on business logic, we might keep it
		bank.PromptPayNumber = existing.PromptPayNumber
	}

	return u.repo.Update(bank)
}

func (u *municipalityBankUseCase) DeleteBank(id uuid.UUID) error {
	return u.repo.Delete(id)
}
