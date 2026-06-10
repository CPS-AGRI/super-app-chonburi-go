package domain

import (
	"time"

	"github.com/google/uuid"
)

type PublicRelation struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	ModuleId      uuid.UUID `gorm:"type:uuid;index;not null;column:module_id" json:"module_id"`
	AdminUserId   uuid.UUID `gorm:"type:uuid;index;not null;column:admin_user_id" json:"admin_user_id"`
	Title         string    `gorm:"type:text;not null;column:title" json:"title"`
	DescriptionTh *string   `gorm:"type:text;column:description_th" json:"description_th"`
	DescriptionEn *string   `gorm:"type:text;column:description_en" json:"description_en"`
	Type          string    `gorm:"type:text;not null;column:type" json:"type"`
	Priority      string    `gorm:"type:text;not null;column:priority" json:"priority"`
	StartDate     time.Time `gorm:"type:timestamptz;not null;column:start_date" json:"start_date"`
	EndDate       time.Time `gorm:"type:timestamptz;not null;column:end_date" json:"end_date"`
	Status        string    `gorm:"type:text;not null;column:status" json:"status"`
	CreatedDate   time.Time `gorm:"type:timestamptz;not null;column:created_date" json:"created_date"`
	UpdatedDate   time.Time `gorm:"type:timestamptz;not null;column:updated_date" json:"updated_date"`
	CreatedBy     string    `gorm:"type:text;not null;column:created_by" json:"created_by"`
	UpdatedBy     string    `gorm:"type:text;not null;column:updated_by" json:"updated_by"`

	Images       []PublicRelationImage       `gorm:"foreignKey:ModulePublicRelationId" json:"images,omitempty"`
	Likes        []PublicRelationLike        `gorm:"foreignKey:ModulePublicRelationId" json:"likes,omitempty"`
	Comments     []PublicRelationComment     `gorm:"foreignKey:ModulePublicRelationId" json:"comments,omitempty"`
	VisitorCount *PublicRelationVisitorCount `gorm:"foreignKey:ModulePublicRelationId" json:"visitor_count,omitempty"`
	AdminUser    *Admin                      `gorm:"foreignKey:AdminUserId" json:"admin_user,omitempty"`
	Module       *Module                     `gorm:"foreignKey:ModuleId" json:"module,omitempty"`
}

func (PublicRelation) TableName() string { return "module_public_relations" }

type PublicRelationVisitorCount struct {
	ModulePublicRelationId uuid.UUID `gorm:"type:uuid;primaryKey;column:module_public_relation_id" json:"module_public_relation_id"`
	Count                  int       `gorm:"type:int4;not null;column:count" json:"count"`
	CreatedDate            time.Time `gorm:"type:timestamptz;not null;column:created_date" json:"created_date"`
	UpdatedDate            time.Time `gorm:"type:timestamptz;not null;column:updated_date" json:"updated_date"`
	CreatedBy              string    `gorm:"type:text;not null;column:created_by" json:"created_by"`
	UpdatedBy              string    `gorm:"type:text;not null;column:updated_by" json:"updated_by"`
}

func (PublicRelationVisitorCount) TableName() string { return "module_public_relation_visitor_count" }

type PublicRelationNotification struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	ModuleId         uuid.UUID  `gorm:"type:uuid;index;not null;column:module_id" json:"module_id"`
	AdminUserId      uuid.UUID  `gorm:"type:uuid;index;not null;column:admin_user_id" json:"admin_user_id"`
	PublicRelationID *uuid.UUID `gorm:"type:uuid;column:public_relation_id" json:"public_relation_id,omitempty"`
	Title            string     `gorm:"type:text;not null;column:title" json:"title"`
	Description      *string    `gorm:"type:text;column:description" json:"description"`
	SendDate         *time.Time `gorm:"type:timestamptz;column:send_date" json:"send_date"`
	Type             string     `gorm:"type:text;not null;column:type" json:"type"`
	Status           string     `gorm:"type:text;not null;column:status" json:"status"`
	ProcessStatus    string     `gorm:"type:text;not null;column:process_status" json:"process_status"`
	CreatedDate      time.Time  `gorm:"type:timestamptz;not null;column:created_date" json:"created_date"`
	UpdatedDate      time.Time  `gorm:"type:timestamptz;not null;column:updated_date" json:"updated_date"`
	CreatedBy        string     `gorm:"type:text;not null;column:created_by" json:"created_by"`
	UpdatedBy        string     `gorm:"type:text;not null;column:updated_by" json:"updated_by"`

	AdminUser      *Admin          `gorm:"foreignKey:AdminUserId" json:"admin_user,omitempty"`
	Module         *Module         `gorm:"foreignKey:ModuleId" json:"module,omitempty"`
	PublicRelation *PublicRelation `gorm:"foreignKey:PublicRelationID" json:"public_relation,omitempty"`
}

func (PublicRelationNotification) TableName() string { return "module_public_relation_notifications" }

type PublicRelationLike struct {
	ModulePublicRelationId uuid.UUID `gorm:"type:uuid;primaryKey;column:module_public_relation_id" json:"module_public_relation_id"`
	UserId                 uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id" json:"user_id"`
	CreatedDate            time.Time `gorm:"type:timestamptz;not null;column:created_date" json:"created_date"`
	UpdatedDate            time.Time `gorm:"type:timestamptz;not null;column:updated_date" json:"updated_date"`
	CreatedBy              string    `gorm:"type:text;not null;column:created_by" json:"created_by"`
	UpdatedBy              string    `gorm:"type:text;not null;column:updated_by" json:"updated_by"`
}

func (PublicRelationLike) TableName() string { return "module_public_relation_likes" }

type PublicRelationImage struct {
	ID                     uuid.UUID `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	ModulePublicRelationId uuid.UUID `gorm:"type:uuid;index;not null;column:module_public_relation_id" json:"module_public_relation_id"`
	Url                    string    `gorm:"type:text;not null;column:url" json:"url"`
	Sequence               int       `gorm:"type:int4;not null;column:sequence" json:"sequence"`
	CreatedDate            time.Time `gorm:"type:timestamptz;not null;column:created_date" json:"created_date"`
	UpdatedDate            time.Time `gorm:"type:timestamptz;not null;column:updated_date" json:"updated_date"`
	CreatedBy              string    `gorm:"type:text;not null;column:created_by" json:"created_by"`
	UpdatedBy              string    `gorm:"type:text;not null;column:updated_by" json:"updated_by"`
}

func (PublicRelationImage) TableName() string { return "module_public_relation_images" }

type PublicRelationComment struct {
	ID                     uuid.UUID `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	ModulePublicRelationId uuid.UUID `gorm:"type:uuid;index;not null;column:module_public_relation_id" json:"module_public_relation_id"`
	UserId                 uuid.UUID `gorm:"type:uuid;index;not null;column:user_id" json:"user_id"`
	Comment                string    `gorm:"type:text;not null;column:comment" json:"comment"`
	Status                 string    `gorm:"type:text;not null;default:'active';column:status" json:"status"`
	CreatedDate            time.Time `gorm:"type:timestamptz;not null;column:created_date" json:"created_date"`
	UpdatedDate            time.Time `gorm:"type:timestamptz;not null;column:updated_date" json:"updated_date"`
	CreatedBy              string    `gorm:"type:text;not null;column:created_by" json:"created_by"`
	UpdatedBy              string    `gorm:"type:text;not null;column:updated_by" json:"updated_by"`

	User *AppUser `gorm:"foreignKey:UserId;references:ID" json:"user,omitempty"`
}

func (PublicRelationComment) TableName() string { return "module_public_relation_comments" }

type MunicipalityWelcomeScreen struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	ImageUrl    string    `gorm:"type:text;not null;column:image_url" json:"image_url"`
	IsActive    bool      `gorm:"type:boolean;not null;default:false;column:is_active" json:"is_active"`
	Type        string    `gorm:"type:text;not null;column:type" json:"type"`
	CreatedDate time.Time `gorm:"type:timestamptz;not null;column:created_date" json:"created_date"`
	UpdatedDate time.Time `gorm:"type:timestamptz;not null;column:updated_date" json:"updated_date"`
	CreatedBy   string    `gorm:"type:text;not null;column:created_by" json:"created_by"`
	UpdatedBy   string    `gorm:"type:text;not null;column:updated_by" json:"updated_by"`
}

func (MunicipalityWelcomeScreen) TableName() string { return "municipality_welcome_screens" }

type PublicRelationQuery struct {
	PageNumber int
	PageSize   int
	Title      *string
	StartDate  *string
	EndDate    *string
}

type PublicRelationNotificationQuery struct {
	PageNumber int
	PageSize   int
	Title      *string
	StartDate  *string
	EndDate    *string
}

type PaginatedPublicRelationResponse struct {
	Items       []PublicRelation `json:"items"`
	TotalItems  int64            `json:"total_items"`
	PageNumber  int              `json:"page_number"`
	PageSize    int              `json:"page_size"`
	TotalPages  int              `json:"total_pages"`
	HasNext     bool             `json:"has_next"`
	HasPrevious bool             `json:"has_previous"`
}

type PaginatedNotificationResponse struct {
	Items       []PublicRelationNotification `json:"items"`
	TotalItems  int64                        `json:"total_items"`
	PageNumber  int                          `json:"page_number"`
	PageSize    int                          `json:"page_size"`
	TotalPages  int                          `json:"total_pages"`
	HasNext     bool                         `json:"has_next"`
	HasPrevious bool                         `json:"has_previous"`
}

type PublicRelationDashboardStats struct {
	ActiveNewsCount        int64 `json:"active_news_count"`
	AverageViewers         int64 `json:"average_viewers"`
	TotalNewsCount         int64 `json:"total_news_count"`
	TotalNotificationCount int64 `json:"total_notification_count"`
	TotalLikesCount        int64 `json:"total_likes_count"`
	TotalCommentsCount     int64 `json:"total_comments_count"`
	ReportedCommentsCount  int64 `json:"reported_comments_count"`
}

type PublicRelationRepository interface {
	GetDashboardStats(moduleId string) (*PublicRelationDashboardStats, error)
	GetPopularNews(moduleId string, limit int) ([]PublicRelation, error)
	GetExpiringNews(moduleId string, limit int) ([]PublicRelation, error)

	GetPaginated(moduleId string, query PublicRelationQuery) (*PaginatedPublicRelationResponse, error)
	GetByID(moduleId string, id string) (*PublicRelation, error)
	Create(pr *PublicRelation) error
	Update(pr *PublicRelation) error
	Delete(moduleId string, id string) error
	HideComment(moduleId string, prId string, commentId string) error
	ShowComment(moduleId string, prId string, commentId string) error

	GetPaginatedNotifications(moduleId string, query PublicRelationNotificationQuery, history bool) (*PaginatedNotificationResponse, error)
	GetNotificationByID(moduleId string, id string) (*PublicRelationNotification, error)
	CreateNotification(notification *PublicRelationNotification) error
	UpdateNotification(notification *PublicRelationNotification) error
	DeleteNotification(moduleId string, id string) error

	GetWelcomeScreens() ([]MunicipalityWelcomeScreen, error)
	CreateWelcomeScreen(screen *MunicipalityWelcomeScreen) error
	UpdateWelcomeScreen(screen *MunicipalityWelcomeScreen) error
	DeleteWelcomeScreen(id string) error
}

type PublicRelationUseCase interface {
	GetDashboardStats(moduleId string) (*PublicRelationDashboardStats, error)
	GetPopularNews(moduleId string, limit int) ([]PublicRelation, error)
	GetExpiringNews(moduleId string, limit int) ([]PublicRelation, error)

	GetPaginated(moduleId string, query PublicRelationQuery) (*PaginatedPublicRelationResponse, error)
	GetByID(moduleId string, id string) (*PublicRelation, error)
	Create(pr *PublicRelation, adminID string) error
	Update(pr *PublicRelation, adminID string) error
	Delete(moduleId string, id string, adminID string) error
	HideComment(moduleId string, prId string, commentId string, adminID string) error
	ShowComment(moduleId string, prId string, commentId string, adminID string) error

	GetPaginatedNotifications(moduleId string, query PublicRelationNotificationQuery, history bool) (*PaginatedNotificationResponse, error)
	GetNotificationByID(moduleId string, id string) (*PublicRelationNotification, error)
	CreateNotification(notification *PublicRelationNotification, adminID string) error
	UpdateNotification(notification *PublicRelationNotification, adminID string) error
	DeleteNotification(moduleId string, id string, adminID string) error

	GetWelcomeScreens() ([]MunicipalityWelcomeScreen, error)
	UploadWelcomeScreen(screen *MunicipalityWelcomeScreen, adminID string) error
}
