package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserActivityTracking struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:id" json:"id"`
	Date      time.Time `gorm:"type:date;column:date" json:"date"`
	ModuleId  uuid.UUID `gorm:"type:uuid;column:module_id" json:"moduleId"`
	ViewCount int       `gorm:"column:view_count;default:0" json:"viewCount"`

	Module *Module `gorm:"foreignKey:ModuleId;references:Id" json:"module,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime;column:created_date" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_date" json:"updatedAt"`
}

func (UserActivityTracking) TableName() string {
	return "user_activity_trackings"
}
