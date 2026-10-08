package usecase

import (
	"context"
	"errors"
	"strings"

	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WeatherAccountUseCase interface {
	GetAll(ctx context.Context) ([]domain.WeatherAccount, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.WeatherAccount, error)
	Create(ctx context.Context, input domain.CreateWeatherAccountInput) (*domain.WeatherAccount, error)
	Update(ctx context.Context, id uuid.UUID, input domain.UpdateWeatherAccountInput) (*domain.WeatherAccount, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type weatherAccountUseCase struct {
	repo repository.WeatherAccountRepository
}

func NewWeatherAccountUseCase(repo repository.WeatherAccountRepository) WeatherAccountUseCase {
	return &weatherAccountUseCase{repo: repo}
}

func (u *weatherAccountUseCase) GetAll(ctx context.Context) ([]domain.WeatherAccount, error) {
	return u.repo.FindAll(ctx)
}

func (u *weatherAccountUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.WeatherAccount, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *weatherAccountUseCase) Create(ctx context.Context, input domain.CreateWeatherAccountInput) (*domain.WeatherAccount, error) {
	input.UniqueID = strings.TrimSpace(input.UniqueID)
	if input.UniqueID == "" {
		return nil, errors.New("unique_id is required")
	}

	// Check if already exists
	existing, err := u.repo.FindByUniqueID(ctx, input.UniqueID)
	if err == nil && existing != nil {
		return nil, errors.New("unique_id already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	isForecastNoti := true
	if input.IsForecastNoti != nil {
		isForecastNoti = *input.IsForecastNoti
	}

	account := &domain.WeatherAccount{
		UniqueID:       input.UniqueID,
		AccountName:    input.AccountName,
		Description:    input.Description,
		IsActive:       isActive,
		IsForecastNoti: isForecastNoti,
	}

	if err := u.repo.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (u *weatherAccountUseCase) Update(ctx context.Context, id uuid.UUID, input domain.UpdateWeatherAccountInput) (*domain.WeatherAccount, error) {
	account, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.AccountName != nil {
		account.AccountName = strings.TrimSpace(*input.AccountName)
	}
	if input.Description != nil {
		account.Description = *input.Description
	}
	if input.IsActive != nil {
		account.IsActive = *input.IsActive
	}
	if input.IsForecastNoti != nil {
		account.IsForecastNoti = *input.IsForecastNoti
	}

	if err := u.repo.Update(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (u *weatherAccountUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	return u.repo.Delete(ctx, id)
}
