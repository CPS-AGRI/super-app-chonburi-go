package repository

import (
	"super-app-chonburi-go/internal/domain"

	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) domain.AdminRefreshTokenRepository {
	return &refreshTokenRepository{db}
}

func (r *refreshTokenRepository) Create(token *domain.AdminRefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) GetByToken(token string) (*domain.AdminRefreshToken, error) {
	var rt domain.AdminRefreshToken
	err := r.db.Where("token = ?", token).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&domain.AdminRefreshToken{}).Error
}

func (r *refreshTokenRepository) DeleteByAdminID(adminID string) error {
	return r.db.Where("admin_user_id = ?", adminID).Delete(&domain.AdminRefreshToken{}).Error
}
