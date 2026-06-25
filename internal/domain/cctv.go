package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CCTV represents a CCTV camera device.
type CCTV struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:id" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null;column:name;index:idx_cctv_name" json:"name"`
	Address     string         `gorm:"type:text;column:address" json:"address"`
	Latitude    float64        `gorm:"type:double precision;not null;column:latitude;index:idx_cctv_lat_lng" json:"latitude"`
	Longitude   float64        `gorm:"type:double precision;not null;column:longitude;index:idx_cctv_lat_lng" json:"longitude"`
	StreamURL   string         `gorm:"type:text;column:stream_url" json:"stream_url"`
	Status      string         `gorm:"type:varchar(20);not null;default:'ONLINE';column:status;index:idx_cctv_status" json:"status"` // 'ONLINE', 'OFFLINE'
	AccessLevel string         `gorm:"type:varchar(20);not null;default:'PUBLIC';column:access_level;index:idx_cctv_access" json:"access_level"` // 'PUBLIC', 'STAFF_ONLY'
	SnapshotURL string         `gorm:"type:text;column:snapshot_url" json:"snapshot_url"`
	CreatorID      *string        `gorm:"type:varchar(50);column:created_by" json:"created_by,omitempty"`
	DeleterID      *string        `gorm:"type:varchar(50);column:deleted_by" json:"deleted_by,omitempty"`
	Creator        *Admin         `gorm:"foreignKey:CreatorID;references:ID" json:"-"`
	Deleter        *Admin         `gorm:"foreignKey:DeleterID;references:ID" json:"-"`
	CreatedAt      time.Time      `gorm:"type:timestamptz;not null;default:now();column:created_at" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"type:timestamptz;not null;default:now();column:updated_at" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
}

// TableName sets the table name for GORM.
func (CCTV) TableName() string {
	return "module_cctv"
}

// CCTVRequest represents a user request to view or get footage from a CCTV camera.
type CCTVRequest struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:id" json:"id"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null;column:user_id;index:idx_cctv_req_user" json:"user_id"`
	CCTVID          uuid.UUID      `gorm:"type:uuid;not null;column:cctv_id;index:idx_cctv_req_cctv" json:"cctv_id"`
	IncidentDate    time.Time      `gorm:"type:date;not null;column:incident_date" json:"incident_date"`
	StartTime       string         `gorm:"type:varchar(10);not null;column:start_time" json:"start_time"`
	EndTime         string         `gorm:"type:varchar(10);not null;column:end_time" json:"end_time"`
	Reason          string         `gorm:"type:text;not null;column:reason" json:"reason"`
	EvidenceFileURL string         `gorm:"type:text;not null;column:evidence_file_url" json:"evidence_file_url"`
	Status          string         `gorm:"type:varchar(20);not null;default:'PENDING';column:status;index:idx_cctv_req_status" json:"status"` // 'PENDING', 'PROCESSING', 'APPROVED', 'REJECTED'
	ApprovedByID    *uuid.UUID     `gorm:"type:uuid;column:approved_by_id" json:"approved_by_id"`
	ResponseFileURL *string        `gorm:"type:text;column:response_file_url" json:"response_file_url"`
	RejectReason    *string        `gorm:"type:text;column:reject_reason" json:"reject_reason"`
	CreatedAt       time.Time      `gorm:"type:timestamptz;not null;default:now();column:created_at" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"type:timestamptz;not null;default:now();column:updated_at" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	// Relations
	CCTV CCTV     `gorm:"foreignKey:CCTVID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"cctv,omitempty"`
	User *AppUser `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName sets the table name for GORM.
func (CCTVRequest) TableName() string {
	return "module_cctv_requests"
}

// CCTVQuery lists query filters for cameras.
type CCTVQuery struct {
	PageNumber int    `json:"page_number"`
	PageSize   int    `json:"page_size"`
	Name       string `json:"name"`
}

// PaginatedCCTVResponse wraps CCTV list response.
type PaginatedCCTVResponse struct {
	Items      []CCTV `json:"items"`
	TotalItems int64  `json:"total_items"`
	PageNumber int    `json:"page_number"`
	TotalPages int    `json:"total_pages"`
}

// CCTVRequestQuery lists query filters for requests.
type CCTVRequestQuery struct {
	PageNumber int `json:"page_number"`
	PageSize   int `json:"page_size"`
}

// PaginatedCCTVRequestResponse wraps CCTVRequest list response.
type PaginatedCCTVRequestResponse struct {
	Items      []CCTVRequest `json:"items"`
	TotalItems int64         `json:"total_items"`
	PageNumber int           `json:"page_number"`
	TotalPages int           `json:"total_pages"`
}

// AdminCCTVRepository defines the data layer interface for Admin CCTV API.
type AdminCCTVRepository interface {
	Create(cctv *CCTV) error
	GetPaginated(query CCTVQuery) (*PaginatedCCTVResponse, error)
	GetRequestsPaginated(query CCTVRequestQuery) (*PaginatedCCTVRequestResponse, error)
	GetRequestByID(id uuid.UUID) (*CCTVRequest, error)
	UpdateRequest(req *CCTVRequest) error
	Delete(id uuid.UUID, adminID string) error
}

// AdminCCTVUseCase defines the business layer interface for Admin CCTV API.
type AdminCCTVUseCase interface {
	CreateCCTV(cctv *CCTV) error
	GetCCTVs(query CCTVQuery) (*PaginatedCCTVResponse, error)
	GetCCTVRequests(query CCTVRequestQuery) (*PaginatedCCTVRequestResponse, error)
	ApproveRequest(id uuid.UUID, responseFileURL string, approvedBy uuid.UUID) error
	RejectRequest(id uuid.UUID, reason string, approvedBy uuid.UUID) error
	DeleteCCTV(id uuid.UUID, adminID string) error
}
