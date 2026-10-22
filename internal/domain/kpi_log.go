package domain

import (
	"time"

	"github.com/google/uuid"
)

// Module Key Constants
const (
	ModuleKeyComplaint       = "complaint"
	ModuleKeyTaxOilGas       = "tax_oil_gas"
	ModuleKeyTaxTobacco      = "tax_tobacco"
	ModuleKeyTaxHotel        = "tax_hotel"
	ModuleKeyPublicRelations = "public_relations"
	ModuleKeyUserVerify      = "user_verify"
	ModuleKeyCCTV            = "cctv"
	ModuleKeyWeather         = "weather"
	ModuleKeyFlood           = "flood"
)

// Complaint Event Codes
const (
	EventComplaintSubmitted        = "COMPLAINT_SUBMITTED"
	EventComplaintCenterForwarded  = "COMPLAINT_CENTER_FORWARDED"
	EventComplaintDeptReceived     = "COMPLAINT_DEPT_RECEIVED"
	EventComplaintManagerAssigned  = "COMPLAINT_MANAGER_ASSIGNED"
	EventComplaintStaffAcknowledged= "COMPLAINT_STAFF_ACKNOWLEDGED"
	EventComplaintStaffAction      = "COMPLAINT_STAFF_ACTION"
	EventComplaintStaffResolved    = "COMPLAINT_STAFF_RESOLVED"
	EventComplaintManagerApproved  = "COMPLAINT_MANAGER_APPROVED"
	EventComplaintStaffRejected    = "COMPLAINT_STAFF_REJECTED"
	EventComplaintDeptReturned     = "COMPLAINT_DEPT_RETURNED"
	EventComplaintCitizenRated     = "COMPLAINT_CITIZEN_RATED"
	EventComplaintCitizenDisputed  = "COMPLAINT_CITIZEN_DISPUTED"
	EventComplaintSLABreached      = "COMPLAINT_SLA_BREACHED"
)

// Public Relations Event Codes
const (
	EventPRNewsCreated           = "PR_NEWS_CREATED"
	EventPRNewsPublished         = "PR_NEWS_PUBLISHED"
	EventPRNewsUpdated           = "PR_NEWS_UPDATED"
	EventPRNewsDeleted           = "PR_NEWS_DELETED"
	EventPRNotificationBroadcast = "PR_NOTIFICATION_BROADCAST"
	EventPRCommentModerated      = "PR_COMMENT_MODERATED"
)

// KPILog represents a record in the kpi_activity_logs table
type KPILog struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"id"`
	ModuleKey             string     `gorm:"type:varchar(50);not null;column:module_key;index" json:"module_key"`
	ModuleTypeID          *uuid.UUID `gorm:"type:uuid;column:module_type_id" json:"module_type_id,omitempty"`
	DepartmentID          *uuid.UUID `gorm:"type:uuid;column:department_id;index" json:"department_id,omitempty"`
	EntityID              uuid.UUID  `gorm:"type:uuid;not null;column:entity_id;index" json:"entity_id"`
	DocumentNumber        *string    `gorm:"type:varchar(100);column:document_number" json:"document_number,omitempty"`
	ActorID               uuid.UUID  `gorm:"type:uuid;not null;column:actor_id;index" json:"actor_id"`
	ActorRole             string     `gorm:"type:varchar(50);not null;column:actor_role" json:"actor_role"`
	TargetUserID          *uuid.UUID `gorm:"type:uuid;column:target_user_id;index" json:"target_user_id,omitempty"`
	TargetRole            *string    `gorm:"type:varchar(50);column:target_role" json:"target_role,omitempty"`
	EventCode             string     `gorm:"type:varchar(60);not null;column:event_code;index" json:"event_code"`
	FromStatus            *string    `gorm:"type:varchar(50);column:from_status" json:"from_status,omitempty"`
	ToStatus              *string    `gorm:"type:varchar(50);column:to_status" json:"to_status,omitempty"`
	StepDurationSeconds   int64      `gorm:"column:step_duration_seconds;default:0" json:"step_duration_seconds"`
	SLATargetSeconds      int64      `gorm:"column:sla_target_seconds;default:0" json:"sla_target_seconds"`
	IsSLABreached         bool       `gorm:"column:is_sla_breached;default:false;index" json:"is_sla_breached"`
	BreachDurationSeconds int64      `gorm:"column:breach_duration_seconds;default:0" json:"breach_duration_seconds"`
	Latitude              *float64   `gorm:"type:numeric(10,7);column:latitude" json:"latitude,omitempty"`
	Longitude             *float64   `gorm:"type:numeric(10,7);column:longitude" json:"longitude,omitempty"`
	Metadata              string     `gorm:"type:jsonb;column:metadata;default:'{}'" json:"metadata"`
	CreatedDate           time.Time  `gorm:"type:timestamptz;not null;default:now();column:created_date;index" json:"created_date"`
}

func (KPILog) TableName() string {
	return "kpi_activity_logs"
}

// KPIThresholdConfig defines dynamic SLA thresholds configured by Executives
type KPIThresholdConfig struct {
	ID                          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"id"`
	ModuleKey                   string    `gorm:"type:varchar(50);not null;uniqueIndex;column:module_key" json:"module_key"`
	ResolutionSLAHours          float64   `gorm:"column:resolution_sla_hours;default:72;not null" json:"resolution_sla_hours"`
	DispatchSLAHours            float64   `gorm:"column:dispatch_sla_hours;default:24;not null" json:"dispatch_sla_hours"`
	ForwardSLAHours             float64   `gorm:"column:forward_sla_hours;default:24;not null" json:"forward_sla_hours"`
	ManagerHighCompletionRate   float64   `gorm:"column:manager_high_completion_rate;default:80;not null" json:"manager_high_completion_rate"`
	ManagerMediumCompletionRate float64   `gorm:"column:manager_medium_completion_rate;default:50;not null" json:"manager_medium_completion_rate"`
	SLAExcellentRate            float64   `gorm:"column:sla_excellent_rate;default:90;not null" json:"sla_excellent_rate"`
	SLAGoodRate                 float64   `gorm:"column:sla_good_rate;default:80;not null" json:"sla_good_rate"`
	UpdatedBy                   string    `gorm:"type:varchar(100);column:updated_by" json:"updated_by"`
	UpdatedDate                 time.Time `gorm:"type:timestamptz;not null;default:now();column:updated_date" json:"updated_date"`
}

func (KPIThresholdConfig) TableName() string {
	return "kpi_threshold_configs"
}

type UpdateKPIThresholdRequest struct {
	ModuleKey                   string  `json:"module_key"`
	ResolutionSLAHours          float64 `json:"resolution_sla_hours"`
	DispatchSLAHours            float64 `json:"dispatch_sla_hours"`
	ForwardSLAHours             float64 `json:"forward_sla_hours"`
	ManagerHighCompletionRate   float64 `json:"manager_high_completion_rate"`
	ManagerMediumCompletionRate float64 `json:"manager_medium_completion_rate"`
	SLAExcellentRate            float64 `json:"sla_excellent_rate"`
	SLAGoodRate                 float64 `json:"sla_good_rate"`
}

// KPI Query & Summary Types for Executive Dashboard
type KPIQueryFilter struct {
	ModuleKey     string     `json:"module_key"`
	DepartmentID  string     `json:"department_id"`
	DepartmentIDs []string   `json:"department_ids"`
	UserID        string     `json:"user_id"`
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
}

type KPISummaryResponse struct {
	TotalEvents           int64               `json:"total_events"`
	TotalResolvedCases    int64               `json:"total_resolved_cases"`
	TotalDispatched       int64               `json:"total_dispatched"`
	AvgMTTRHours          float64             `json:"avg_mttr_hours"`
	AvgMTTRDays           float64             `json:"avg_mttr_days"`
	AvgDispatchHours      float64             `json:"avg_dispatch_hours"`
	SLAComplianceRate     float64             `json:"sla_compliance_rate"`
	TotalSLABreachedCases int64               `json:"total_sla_breached_cases"`
	AvgCitizenRating      float64             `json:"avg_citizen_rating"`
	TotalDisputes         int64               `json:"total_disputes"`
	Thresholds            *KPIThresholdConfig `json:"thresholds,omitempty"`

	// Public Relations KPI Metrics (ModuleKey == public_relations)
	TotalPublishedNews  int64   `json:"total_published_news,omitempty"`
	TotalNewsViews      int64   `json:"total_news_views,omitempty"`
	AvgViewsPerNews     float64 `json:"avg_views_per_news,omitempty"`
	TotalNewsLikes      int64   `json:"total_news_likes,omitempty"`
	TotalNewsComments   int64   `json:"total_news_comments,omitempty"`
	TotalPushBroadcasts int64   `json:"total_push_broadcasts,omitempty"`
	EngagementRate      float64 `json:"engagement_rate,omitempty"`
}

type KPIStaffLeaderboardItem struct {
	UserID            string  `json:"user_id"`
	Name              string  `json:"name"`
	LastName          string  `json:"last_name"`
	DepartmentName    string  `json:"department_name"`
	RoleType          string  `json:"role_type"`
	Position          string  `json:"position"`
	TotalAssigned     int64   `json:"total_assigned"`
	TotalResolved     int64   `json:"total_resolved"`
	AvgMTTRHours      float64 `json:"avg_mttr_hours"`
	SLAComplianceRate float64 `json:"sla_compliance_rate"`
	AvgRating         float64 `json:"avg_rating"`
	DisputeCount      int64   `json:"dispute_count"`

	// PR Specific Staff Metrics
	PublishedNewsCount    int64 `json:"published_news_count,omitempty"`
	TotalViewsGenerated   int64 `json:"total_views_generated,omitempty"`
	TotalLikesReceived    int64 `json:"total_likes_received,omitempty"`
	TotalCommentsReceived int64 `json:"total_comments_received,omitempty"`
	PushBroadcastCount    int64 `json:"push_broadcast_count,omitempty"`
}

type KPIManagerScorecardItem struct {
	ManagerID                   string  `json:"manager_id"`
	Name                        string  `json:"name"`
	LastName                    string  `json:"last_name"`
	DepartmentID                string  `json:"department_id"`
	DepartmentName              string  `json:"department_name"`
	RoleType                    string  `json:"role_type"`
	Position                    string  `json:"position"`
	TotalDispatched             int64   `json:"total_dispatched"`
	TotalResolvedFromDispatched int64   `json:"total_resolved_from_dispatched"`
	CompletionRate              float64 `json:"completion_rate"`
	AvgDispatchHours            float64 `json:"avg_dispatch_hours"`
	UnassignedCount             int64   `json:"unassigned_count"`
	RejectedReturnCount         int64   `json:"rejected_return_count"`
}

type KPIPersonalCaseItem struct {
	EntityID            string    `json:"entity_id"`
	DocumentNumber      string    `json:"document_number"`
	EventCode           string    `json:"event_code"`
	FromStatus          string    `json:"from_status"`
	ToStatus            string    `json:"to_status"`
	StepDurationSeconds int64     `json:"step_duration_seconds"`
	SLATargetSeconds    int64     `json:"sla_target_seconds"`
	IsSLABreached       bool      `json:"is_sla_breached"`
	BreachDuration      int64     `json:"breach_duration_seconds"`
	CreatedDate         time.Time `json:"created_date"`
	Metadata            string    `json:"metadata"`
}

type KPIStaffDetailResponse struct {
	UserID            string                `json:"user_id"`
	Name              string                `json:"name"`
	LastName          string                `json:"last_name"`
	DepartmentName    string                `json:"department_name"`
	TotalAssigned     int64                 `json:"total_assigned"`
	TotalResolved     int64                 `json:"total_resolved"`
	AvgMTTRHours      float64               `json:"avg_mttr_hours"`
	SLAComplianceRate float64               `json:"sla_compliance_rate"`
	TotalSLABreached  int64                 `json:"total_sla_breached"`
	AvgRating         float64               `json:"avg_rating"`
	DisputeCount      int64                 `json:"dispute_count"`
	RecentCases       []KPIPersonalCaseItem `json:"recent_cases"`
}

type KPIManagerDispatchCaseItem struct {
	EntityID            string    `json:"entity_id"`
	DocumentNumber      string    `json:"document_number"`
	Title               string    `json:"title"`
	CategoryName        string    `json:"category_name"`
	TargetUserName      string    `json:"target_user_name"`
	DepartmentName      string    `json:"department_name"`
	EventCode           string    `json:"event_code"`
	StepDurationSeconds int64     `json:"step_duration_seconds"`
	SLATargetSeconds    int64     `json:"sla_target_seconds"`
	IsSLABreached       bool      `json:"is_sla_breached"`
	BreachDuration      int64     `json:"breach_duration_seconds"`
	CreatedDate         time.Time `json:"created_date"`
	Reason              string    `json:"reason"`
}

type KPIManagerDetailResponse struct {
	ManagerID                   string                       `json:"manager_id"`
	Name                        string                       `json:"name"`
	LastName                    string                       `json:"last_name"`
	DepartmentName              string                       `json:"department_name"`
	Position                    string                       `json:"position"`
	TotalDispatched             int64                        `json:"total_dispatched"`
	TotalResolvedFromDispatched int64                        `json:"total_resolved_from_dispatched"`
	CompletionRate              float64                      `json:"completion_rate"`
	AvgDispatchHours            float64                      `json:"avg_dispatch_hours"`
	SLAComplianceRate           float64                      `json:"sla_compliance_rate"`
	TotalSLABreached            int64                        `json:"total_sla_breached"`
	RejectedReturnCount         int64                        `json:"rejected_return_count"`
	RecentCases                 []KPIManagerDispatchCaseItem `json:"recent_cases"`
	Thresholds                  *KPIThresholdConfig          `json:"thresholds,omitempty"`
}

type KPICategoryStatItem struct {
	ModuleTypeID  string  `json:"module_type_id"`
	CategoryName  string  `json:"category_name"`
	TotalCount    int64   `json:"total_count"`
	ResolvedCount int64   `json:"resolved_count"`
	PendingCount  int64   `json:"pending_count"`
	Percentage    float64 `json:"percentage"`
}

type KPIBreachedCaseItem struct {
	ID               string     `json:"id"`
	DocumentNumber   string     `json:"document_number"`
	Title            string     `json:"title"`
	CategoryName     string     `json:"category_name"`
	DepartmentID     string     `json:"department_id"`
	DepartmentName   string     `json:"department_name"`
	AssigneeName     string     `json:"assignee_name"`
	AssignerName     string     `json:"assigner_name"`
	Status           string     `json:"status"`
	CreatedDate      time.Time  `json:"created_date"`
	ResolvedDate     *time.Time `json:"resolved_date,omitempty"`
	DurationSeconds  int64      `json:"duration_seconds"`
	SLATargetSeconds int64      `json:"sla_target_seconds"`
	OverdueSeconds   int64      `json:"overdue_seconds"`
	IsActive         bool       `json:"is_active"`
}

// KPILogRepository interface
type KPILogRepository interface {
	Create(log *KPILog) error
	CreateBatch(logs []KPILog) error
	GetKPISummary(filter KPIQueryFilter) (*KPISummaryResponse, error)
	GetStaffLeaderboard(filter KPIQueryFilter) ([]KPIStaffLeaderboardItem, error)
	GetManagerScorecard(filter KPIQueryFilter) ([]KPIManagerScorecardItem, error)
	GetStaffDetail(userID string, filter KPIQueryFilter) (*KPIStaffDetailResponse, error)
	GetManagerDetail(userID string, filter KPIQueryFilter) (*KPIManagerDetailResponse, error)
	GetBreachedCases(filter KPIQueryFilter) ([]KPIBreachedCaseItem, error)
	GetCategoryStats(filter KPIQueryFilter) ([]KPICategoryStatItem, error)
	GetThresholds(moduleKey string) (*KPIThresholdConfig, error)
	UpdateThresholds(moduleKey string, req UpdateKPIThresholdRequest, adminID string) (*KPIThresholdConfig, error)
}

// KPILogger represents an async logging engine
type KPILogger interface {
	LogEvent(log *KPILog)
	Close()
}

// KPIUseCase interface
type KPIUseCase interface {
	GetSummary(filter KPIQueryFilter) (*KPISummaryResponse, error)
	GetStaffLeaderboard(filter KPIQueryFilter) ([]KPIStaffLeaderboardItem, error)
	GetManagerScorecard(filter KPIQueryFilter) ([]KPIManagerScorecardItem, error)
	GetStaffDetail(userID string, filter KPIQueryFilter) (*KPIStaffDetailResponse, error)
	GetManagerDetail(userID string, filter KPIQueryFilter) (*KPIManagerDetailResponse, error)
	GetBreachedCases(filter KPIQueryFilter) ([]KPIBreachedCaseItem, error)
	GetCategoryStats(filter KPIQueryFilter) ([]KPICategoryStatItem, error)
	GetThresholds(moduleKey string) (*KPIThresholdConfig, error)
	UpdateThresholds(moduleKey string, req UpdateKPIThresholdRequest, adminID string) (*KPIThresholdConfig, error)
}
