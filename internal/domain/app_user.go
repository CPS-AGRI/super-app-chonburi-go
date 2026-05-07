package domain

import (
	"time"

	"github.com/google/uuid"
)

type AppUserStatus string

const (
	AppUserStatusPending  AppUserStatus = "pending"
	AppUserStatusActive   AppUserStatus = "active"
	AppUserStatusInactive AppUserStatus = "inactive"
)

// AppUser maps to the "users" table.
// Name changed from User to AppUser to avoid conflict with the User interface in admin.go.
type AppUser struct {
	// Core Authentication (As requested)
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:id" json:"id"`
	PhoneNumber     string    `gorm:"type:text;not null;column:phone_number" json:"phone_number"`
	PinHash         string    `gorm:"type:text;not null;column:pin_hash" json:"-"`
	IsConsent       bool      `gorm:"not null;default:false;column:is_consent" json:"is_consent"`
	ImageProfileUrl *string   `gorm:"type:text;column:image_profile_url" json:"image_profile_url"`

	// Future-proofing: OAuth & External ID Support
	ThaiId     *string `gorm:"type:text;column:thai_id" json:"thai_id"`         // เลขบัตรประชาชน
	Provider   *string `gorm:"type:text;column:provider" json:"provider"`       // e.g., "local", "google", "line", "facebook"
	ProviderId *string `gorm:"type:text;column:provider_id" json:"provider_id"` // OAuth UID
	Email      *string `gorm:"type:text;column:email" json:"email"`

	// Profile Information (Required for Dashboard Analytics)
	Prefix         *string       `gorm:"type:text;column:prefix" json:"prefix"`
	Name           *string       `gorm:"type:text;column:name" json:"name"`
	LastName       *string       `gorm:"type:text;column:last_name" json:"last_name"`
	Birthday       *time.Time    `gorm:"type:date;column:birthday" json:"birthday"`
	Status         AppUserStatus `gorm:"type:text;default:'pending';column:status" json:"status"`

	// Auditing (As requested)
	CreatedBy   string    `gorm:"type:text;not null;default:'';column:created_by" json:"created_by"`
	CreatedDate time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;column:created_date" json:"created_date"`
	UpdatedBy   string    `gorm:"type:text;not null;default:'';column:updated_by" json:"updated_by"`
	UpdatedDate time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;column:updated_date" json:"updated_date"`
}

func (AppUser) TableName() string {
	return "users"
}
