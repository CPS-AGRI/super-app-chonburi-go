package domain

import (
	"time"
	"github.com/google/uuid"
)

type ActivityLog struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RequestTime        time.Time `json:"requestTime"`
	AdminName          string    `json:"adminName"`
	Method             string    `json:"method"`
	Path               string    `json:"path"`
	ResponseStatusCode int       `json:"responseStatusCode"`
	DurationMs         int64     `json:"durationMs"`
	AdminID            string    `json:"adminId"`
	IPAddress          string    `json:"ipAddress"`
	UserAgent          string    `json:"userAgent"`
	TraceID            string    `json:"traceId"`
	ResponseTime       time.Time `json:"responseTime"`
}

type ActivityLogRepository interface {
	Create(log *ActivityLog) error
}
