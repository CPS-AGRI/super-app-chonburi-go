package repository

import (
	"context"
	"time"

	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/redis"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WeatherAccountRepository interface {
	FindAll(ctx context.Context) ([]domain.WeatherAccount, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.WeatherAccount, error)
	FindByUniqueID(ctx context.Context, uniqueID string) (*domain.WeatherAccount, error)
	Create(ctx context.Context, account *domain.WeatherAccount) error
	Update(ctx context.Context, account *domain.WeatherAccount) error
	Delete(ctx context.Context, id uuid.UUID) error
	FlushCache(ctx context.Context)
}

type weatherAccountRepository struct {
	db *gorm.DB
}

func NewWeatherAccountRepository(db *gorm.DB) WeatherAccountRepository {
	return &weatherAccountRepository{db: db}
}

func (r *weatherAccountRepository) FindAll(ctx context.Context) ([]domain.WeatherAccount, error) {
	var items []domain.WeatherAccount
	err := r.db.WithContext(ctx).Order("created_at ASC").Find(&items).Error
	return items, err
}

func (r *weatherAccountRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.WeatherAccount, error) {
	var item domain.WeatherAccount
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *weatherAccountRepository) FindByUniqueID(ctx context.Context, uniqueID string) (*domain.WeatherAccount, error) {
	var item domain.WeatherAccount
	err := r.db.WithContext(ctx).First(&item, "unique_id = ?", uniqueID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *weatherAccountRepository) Create(ctx context.Context, account *domain.WeatherAccount) error {
	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Create(account).Error
	if err == nil {
		r.FlushCache(ctx)
	}
	return err
}

func (r *weatherAccountRepository) Update(ctx context.Context, account *domain.WeatherAccount) error {
	account.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Save(account).Error
	if err == nil {
		r.FlushCache(ctx)
	}
	return err
}

func (r *weatherAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Delete(&domain.WeatherAccount{}, "id = ?", id).Error
	if err == nil {
		r.FlushCache(ctx)
	}
	return err
}

func (r *weatherAccountRepository) FlushCache(ctx context.Context) {
	if redis.Client != nil {
		_ = redis.Client.Del(ctx, "chonburi:weather:stations:all").Err()
	}
}
