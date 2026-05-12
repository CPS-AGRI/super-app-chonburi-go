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
	// Check for time overlap
	allShifts, err := u.GetAllShifts()
	if err == nil {
		for _, s := range allShifts {
			if s.ID == schedule.ID {
				continue // Skip checking against itself when updating
			}
			if s.WorkingDay == schedule.WorkingDay {
				// overlap condition: (StartA < EndB) and (EndA > StartB)
				if schedule.WorkingHoursStart < s.WorkingHoursEnd && schedule.WorkingHoursEnd > s.WorkingHoursStart {
					return domain.ErrScheduleOverlap
				}
			}
		}
	}

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
