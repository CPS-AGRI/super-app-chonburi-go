package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserActivityTracking struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:id" json:"id"`
	Date      time.Time `gorm:"type:date;column:date;not null;index:idx_user_activity_date_module,unique" json:"date"`
	ModuleId  uuid.UUID `gorm:"type:uuid;column:module_id;not null;index:idx_user_activity_date_module,unique" json:"moduleId"`
	ViewCount int       `gorm:"column:view_count;default:0;not null" json:"viewCount"`

	Module *Module `gorm:"foreignKey:ModuleId;references:Id" json:"module,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime;column:created_date;type:timestamptz" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_date;type:timestamptz" json:"updatedAt"`
}

func (UserActivityTracking) TableName() string {
	return "user_activity_trackings"
}
