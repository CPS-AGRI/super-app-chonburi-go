package domain

import (
	"time"

	"github.com/google/uuid"
)

type WeatherAccount struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UniqueID       string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"unique_id"`
	AccountName    string    `gorm:"type:varchar(128);not null;default:''" json:"account_name"`
	Description    string    `gorm:"type:text" json:"description,omitempty"`
	IsActive       bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	IsForecastNoti bool      `gorm:"type:boolean;not null;default:true" json:"is_forecast_noti"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (WeatherAccount) TableName() string {
	return "chonburi_weather_accounts"
}

type CreateWeatherAccountInput struct {
	UniqueID       string `json:"unique_id" validate:"required"`
	AccountName    string `json:"account_name" validate:"required"`
	Description    string `json:"description"`
	IsActive       *bool  `json:"is_active"`
	IsForecastNoti *bool  `json:"is_forecast_noti"`
}

type UpdateWeatherAccountInput struct {
	AccountName    *string `json:"account_name"`
	Description    *string `json:"description"`
	IsActive       *bool   `json:"is_active"`
	IsForecastNoti *bool   `json:"is_forecast_noti"`
}
