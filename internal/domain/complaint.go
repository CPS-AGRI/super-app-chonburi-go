package domain

import (
	"time"
	"github.com/google/uuid"
)

const (
	ComplaintStatusDraft      = "draft"
	ComplaintStatusPending    = "pending"
	ComplaintStatusReceived   = "received"
	ComplaintStatusInProgress = "in_progress"
	ComplaintStatusCompleted  = "completed"
	ComplaintStatusRejected   = "rejected"
)

// module_complaints
type Complaint struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	ModuleTypeId      string         `gorm:"type:uuid;index;not null;column:module_type_id" json:"module_type_id"`
	UserId            uuid.UUID      `gorm:"type:uuid;index;not null;column:user_id" json:"user_id"`
	DocumentId        string         `gorm:"not null;column:document_id;index" json:"document_id"`
	Description       string         `gorm:"type:text;column:description" json:"description"`
	Latitude          float64        `gorm:"column:latitude" json:"latitude"`
	Longitude         float64        `gorm:"column:longitude" json:"longitude"`
	Status            string         `gorm:"not null;column:status;index" json:"status"`
	AssignerId        *string        `gorm:"type:uuid;index;column:assigner_id" json:"assigner_id"`
	AssigneeId        *string        `gorm:"type:uuid;index;column:assignee_id" json:"assignee_id"`
	CreatedDate       time.Time      `gorm:"not null;type:timestamptz;column:created_date;index" json:"created_at"`
	UpdatedDate       time.Time      `gorm:"not null;type:timestamptz;column:updated_date" json:"updated_at"`
	CreatedBy         string         `gorm:"not null;column:created_by" json:"created_by"`
	UpdatedBy         string         `gorm:"not null;column:updated_by" json:"updated_by"`
	DepartmentId      *string        `gorm:"type:uuid;index;column:department_id" json:"department_id"`
	IsOtherModuleType bool           `gorm:"not null;default:false;column:is_other_module_type" json:"is_other_module_type"`

	// Relations
	ModuleType      *ModuleType      `gorm:"foreignKey:ModuleTypeId" json:"module_type,omitempty"`
	Department      *Department      `gorm:"foreignKey:DepartmentId" json:"department,omitempty"`
	User            *AppUser         `gorm:"foreignKey:UserId;references:ID" json:"user,omitempty"`
	UserInformation *UserInformation `gorm:"-" json:"user_information,omitempty"`
	Images          []ComplaintImage `gorm:"foreignKey:ModuleComplaintId" json:"images,omitempty"`
	Activities      []ComplaintActivity `gorm:"foreignKey:ModuleComplaintId" json:"activities,omitempty"`
	Assignee        *Admin           `gorm:"-" json:"assignee,omitempty"`
}

func (Complaint) TableName() string { return "module_complaints" }

// module_complaint_images
type ComplaintImage struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	ModuleComplaintId uuid.UUID `gorm:"type:uuid;index;not null;column:module_complaint_id" json:"module_complaint_id"`
	Url               string    `gorm:"not null;column:url" json:"url"`
	Sequence          int       `gorm:"not null;column:sequence" json:"sequence"`
	CreatedDate       time.Time `gorm:"not null;type:timestamptz;column:created_date" json:"created_at"`
	UpdatedDate       time.Time `gorm:"not null;type:timestamptz;column:updated_date" json:"updated_at"`
	CreatedBy         string    `gorm:"not null;column:created_by" json:"created_by"`
	UpdatedBy         string    `gorm:"not null;column:updated_by" json:"updated_by"`
}

func (ComplaintImage) TableName() string { return "module_complaint_images" }

// module_complaint_activities
type ComplaintActivity struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	ModuleComplaintId uuid.UUID `gorm:"type:uuid;index;not null;column:module_complaint_id" json:"module_complaint_id"`
	Description       string    `gorm:"type:text;column:description" json:"description"`
	Status            string    `gorm:"column:status;index" json:"status"`
	CreatedDate       time.Time `gorm:"not null;type:timestamptz;column:created_date" json:"created_at"`
	UpdatedDate       time.Time `gorm:"not null;type:timestamptz;column:updated_date" json:"updated_at"`
	CreatedBy         string    `gorm:"not null;column:created_by" json:"created_by"`
	UpdatedBy         string    `gorm:"not null;column:updated_by" json:"updated_by"`

	Images []ComplaintActivityImage `gorm:"foreignKey:ModuleComplaintActivityId" json:"images,omitempty"`
	Admin  *Admin                   `gorm:"-" json:"admin,omitempty"`
}

func (ComplaintActivity) TableName() string { return "module_complaint_activities" }

// module_complaint_activity_images
type ComplaintActivityImage struct {
	ID                        uuid.UUID `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	ModuleComplaintActivityId uuid.UUID `gorm:"type:uuid;index;not null;column:module_complaint_activity_id" json:"module_complaint_activity_id"`
	Url                       string    `gorm:"not null;column:url" json:"url"`
	Sequence                  int       `gorm:"not null;column:sequence" json:"sequence"`
	CreatedDate               time.Time `gorm:"not null;type:timestamptz;column:created_date" json:"created_at"`
	UpdatedDate               time.Time `gorm:"not null;type:timestamptz;column:updated_date" json:"updated_at"`
	CreatedBy                 string    `gorm:"not null;column:created_by" json:"created_by"`
	UpdatedBy                 string    `gorm:"not null;column:updated_by" json:"updated_by"`
}

func (ComplaintActivityImage) TableName() string { return "module_complaint_activity_images" }

// Interfaces
type ComplaintQuery struct {
	PageNumber int `query:"page_number"`
	PageSize   int `query:"page_size"`
	Status     []string
	AssigneeId *string `query:"assignee_id"`

	// Search Filters
	IdentityNumber *string `query:"identity_number"`
	Name           *string `query:"name"`
	LastName       *string `query:"last_name"`
	StartDate      *string `query:"start_date"`
	EndDate        *string `query:"end_date"`
	DepartmentID   *string `query:"department_id"`

	AllowedModuleTypeIDs []string
	IsSuperAdmin         bool
	IsComplaintCenter    bool
	AdminDepartmentIDs   []string
	HasBeenAssigned      *bool
}

type PaginatedComplaintResponse struct {
	Items       []Complaint `json:"items"`
	TotalItems  int64       `json:"total_items"`
	PageNumber  int         `json:"page_number"`
	PageSize    int         `json:"page_size"`
	TotalPages  int         `json:"total_pages"`
	HasNext     bool        `json:"has_next"`
	HasPrevious bool        `json:"has_previous"`
}

type ComplaintRepository interface {
	GetPaginated(query ComplaintQuery) (*PaginatedComplaintResponse, error)
	GetByID(id string, allowedModuleTypeIDs []string, isSuperAdmin bool) (*Complaint, error)
	Create(complaint *Complaint) error
	Update(complaint *Complaint) error
	CreateActivity(activity *ComplaintActivity) error
	Delete(id string) error
	GetAllowedModuleTypeIDs(deptIDs []string) ([]string, error)
}

type ComplaintUseCase interface {
	GetComplaints(query ComplaintQuery, adminID string) (*PaginatedComplaintResponse, error)
	GetComplaintByID(id string, adminID string) (*Complaint, error)
	CreateComplaint(complaint *Complaint) error
	UpdateComplaintStatus(id string, status string, description string, adminID string, images []string) error
	ForwardComplaint(id string, departmentID string, description string, adminID string) error
	AssignComplaint(id string, assigneeID string, description string, adminID string) error
	RejectComplaint(id string, reason string, adminID string) error
	AddActivity(activity *ComplaintActivity, adminID string) error
	DeleteComplaint(id string) error
}
