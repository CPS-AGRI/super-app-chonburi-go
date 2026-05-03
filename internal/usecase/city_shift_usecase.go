package usecase

import (
	"errors"
	"time"

	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
)

type cityShiftUseCase struct {
	repo domain.CityShiftRepository
}

func NewCityShiftUseCase(repo domain.CityShiftRepository) domain.CityShiftUseCase {
	return &cityShiftUseCase{repo: repo}
}

func (u *cityShiftUseCase) GetAllShifts() ([]domain.CityShift, error) {
	return u.repo.GetAll()
}

func (u *cityShiftUseCase) SaveShift(shift *domain.CityShift) error {
	// 1. Check for overlapping shifts on the same day
	shiftsOnDay, err := u.repo.GetByDay(shift.WorkingDay)
	if err != nil {
		return err
	}

	for _, existingShift := range shiftsOnDay {
		// Skip self if updating
		if shift.ID != uuid.Nil && existingShift.ID == shift.ID {
			continue
		}

		if timeRangesOverlap(existingShift.WorkingHoursStart, existingShift.WorkingHoursEnd, shift.WorkingHoursStart, shift.WorkingHoursEnd) {
			return errors.New("ไม่สามารถเพิ่มได้เนื่องจากเวลาคาบเกี่ยวกัน")
		}
	}

	// 2. Create or Update
	if shift.ID == uuid.Nil {
		return u.repo.Create(shift)
	}

	// For update, preserve existing metadata
	existing, err := u.repo.GetByID(shift.ID)
	if err != nil {
		return err
	}
	
	shift.Status = existing.Status
	shift.CreatedBy = existing.CreatedBy
	shift.CreatedAt = existing.CreatedAt

	return u.repo.Update(shift)
}

func (u *cityShiftUseCase) DeleteShift(id uuid.UUID) error {
	return u.repo.Delete(id)
}

// timeRangesOverlap checks if two time ranges (HH:MM) overlap
func timeRangesOverlap(start1, end1, start2, end2 string) bool {
	layout := "15:04"
	t1Start, err1 := time.Parse(layout, start1)
	t1End, err2 := time.Parse(layout, end1)
	t2Start, err3 := time.Parse(layout, start2)
	t2End, err4 := time.Parse(layout, end2)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return false
	}

	return t1Start.Before(t2End) && t2Start.Before(t1End)
}
