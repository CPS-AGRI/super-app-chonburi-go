package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Complaint Statuses
const (
	ComplaintStatusReceived   = "Received"
	ComplaintStatusInProgress = "InProgress"
	ComplaintStatusCompleted  = "Completed"
	ComplaintStatusRejected   = "Rejected"
)

// ComplaintUserInformation represents the citizen who submitted the complaint
type ComplaintUserInformation struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ComplaintID    uuid.UUID `gorm:"type:uuid;index;not null;constraint:OnDelete:CASCADE" json:"complaintId"`
	UserID         *string   `json:"userId"` // Optional, for future mobile app linking
	Prefix         string    `json:"prefix"`
	Name           string    `json:"name"`
	LastName       string    `json:"lastName"`
	Phone          string    `json:"phone"`
	IdentityNumber string    `json:"identityNumber"`
}

// Complaint represents a single complaint document
type Complaint struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	DocumentID    string    `gorm:"uniqueIndex;not null" json:"documentId"` // Auto-generated CMP-YYYYMMDD-XXXX
	PermissionID  string    `gorm:"not null;index" json:"permissionId"`     // Maps to SystemPermission.ID (The topic/sub-module)
	Latitude      string    `json:"latitude"`
	Longitude     string    `json:"longitude"`
	GoogleMapsUrl string    `json:"googleMapsUrl"`
	Description   string    `gorm:"type:text" json:"description"`
	Status        string    `gorm:"not null;default:'Received'" json:"status"`

	// Foreign Keys to Admins
	AssignerID *uuid.UUID `gorm:"type:uuid;index" json:"assignerId"`   // The Manager who assigned it
	AssigneeID *uuid.UUID `gorm:"type:uuid;index" json:"assigneeId"`   // The Employee working on it
	RejectedByID *uuid.UUID `gorm:"type:uuid;index" json:"rejectedById"` // The Manager who rejected it

	// Relations
	Assigner   *Admin `gorm:"foreignKey:AssignerID" json:"assigner,omitempty"`
	Assignee   *Admin `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	RejectedBy *Admin `gorm:"foreignKey:RejectedByID" json:"rejectedBy,omitempty"`
	Permission *SystemPermission `gorm:"foreignKey:PermissionID;references:ID" json:"permission,omitempty"`

	UserInformation *ComplaintUserInformation `gorm:"foreignKey:ComplaintID" json:"userInformation,omitempty"`
	Images          []ComplaintImage          `gorm:"foreignKey:ComplaintID;constraint:OnDelete:CASCADE" json:"images,omitempty"`
	Activities      []ComplaintActivity       `gorm:"foreignKey:ComplaintID;constraint:OnDelete:CASCADE" json:"activities,omitempty"`

	CreatedBy string         `json:"createdBy"`
	UpdatedBy string         `json:"updatedBy"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ComplaintImage represents an image attached to the complaint when initially submitted
type ComplaintImage struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ComplaintID uuid.UUID `gorm:"type:uuid;index;not null" json:"complaintId"`
	URL         string    `gorm:"not null" json:"url"`
	Sequence    int       `gorm:"not null;default:0" json:"sequence"`
}

// ComplaintActivity represents an update/action taken on the complaint
type ComplaintActivity struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ComplaintID uuid.UUID `gorm:"type:uuid;index;not null" json:"complaintId"`
	Description string    `gorm:"type:text" json:"description"`
	Status      string    `gorm:"not null" json:"status"` // Status of the complaint at this activity

	// The Admin who made the update (Employee or Manager)
	AdminID *uuid.UUID `gorm:"type:uuid;index" json:"adminId"`
	Admin   *Admin     `gorm:"foreignKey:AdminID" json:"admin,omitempty"`

	Images []ComplaintActivityImage `gorm:"foreignKey:ActivityID;constraint:OnDelete:CASCADE" json:"images,omitempty"`

	CreatedBy string    `json:"createdBy"`
	UpdatedBy string    `json:"updatedBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ComplaintActivityImage represents an image attached to an activity update
type ComplaintActivityImage struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ActivityID uuid.UUID `gorm:"type:uuid;index;not null" json:"activityId"`
	URL        string    `gorm:"not null" json:"url"`
	Sequence   int       `gorm:"not null;default:0" json:"sequence"`
}

// --- Table Names ---

func (ComplaintUserInformation) TableName() string {
	return "module_complaint_user_informations"
}

func (Complaint) TableName() string {
	return "module_complaints"
}

func (ComplaintImage) TableName() string {
	return "module_complaint_images"
}

func (ComplaintActivity) TableName() string {
	return "module_complaint_activities"
}

func (ComplaintActivityImage) TableName() string {
	return "module_complaint_activity_images"
}

// --- Interfaces ---

type ComplaintQuery struct {
	PageNumber int
	PageSize   int
	Status     []string
	AssigneeID *uuid.UUID

	// Internally used for filtering based on logged-in user's department
	AllowedPermissionIDs []string
	IsSuperAdmin         bool
}

type PaginatedComplaintResponse struct {
	Items       []Complaint `json:"items"`
	TotalItems  int64       `json:"totalItems"`
	PageNumber  int         `json:"pageNumber"`
	PageSize    int         `json:"pageSize"`
	TotalPages  int         `json:"totalPages"`
	HasNext     bool        `json:"hasNext"`
	HasPrevious bool        `json:"hasPrevious"`
}

type ComplaintRepository interface {
	GetPaginated(query ComplaintQuery) (*PaginatedComplaintResponse, error)
	GetByID(id uuid.UUID, allowedPermissionIDs []string, isSuperAdmin bool) (*Complaint, error)
	Create(complaint *Complaint) error
	Update(complaint *Complaint) error
	CreateActivity(activity *ComplaintActivity) error
	Delete(id uuid.UUID) error
}

type ComplaintUseCase interface {
	GetComplaints(query ComplaintQuery, adminID uuid.UUID) (*PaginatedComplaintResponse, error)
	GetComplaintByID(id uuid.UUID, adminID uuid.UUID) (*Complaint, error)
	
	// Called by Mock Citizen (or Real Citizen later)
	CreateComplaint(complaint *Complaint) error
	
	// Called by Managers
	AssignComplaint(id uuid.UUID, assignerID uuid.UUID, assigneeID uuid.UUID) error
	RejectComplaint(id uuid.UUID, rejecterID uuid.UUID, reason string) error
	
	// Called by Employees
	AddActivity(activity *ComplaintActivity, adminID uuid.UUID) error
	DeleteComplaint(id uuid.UUID, adminID uuid.UUID) error
}
