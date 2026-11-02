package usecase

import (
	"super-app-chonburi-go/internal/domain"
	"time"

	"github.com/google/uuid"
)

type municipalityWorkScheduleUseCase struct {
	repo domain.MunicipalityWorkScheduleRepository
}

func NewMunicipalityWorkScheduleUseCase(repo domain.MunicipalityWorkScheduleRepository) domain.MunicipalityWorkScheduleUseCase {
	return &municipalityWorkScheduleUseCase{repo: repo}
}

func (u *municipalityWorkScheduleUseCase) GetAllShifts() ([]domain.MunicipalityWorkSchedule, error) {
	return u.repo.GetAll()
}

func (u *municipalityWorkScheduleUseCase) SaveShift(schedule *domain.MunicipalityWorkSchedule) error {
	if schedule.ID == uuid.Nil {
		schedule.ID = uuid.New()
		schedule.CreatedAt = time.Now()
		schedule.UpdatedAt = time.Now()
		return u.repo.Create(schedule)
	}

	existing, err := u.repo.GetByID(schedule.ID)
	if err != nil {
		return err
	}

	schedule.CreatedAt = existing.CreatedAt
	schedule.UpdatedAt = time.Now()

	return u.repo.Update(schedule)
}

func (u *municipalityWorkScheduleUseCase) DeleteShift(id uuid.UUID) error {
	return u.repo.Delete(id)
}
