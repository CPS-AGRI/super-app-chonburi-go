package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminRole represents the hierarchy/title (e.g., SuperAdmin, Head, Staff)
type AdminRole struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name         string         `gorm:"unique;not null" json:"name"`
	Description  *string        `json:"description"`
	IsSuperAdmin bool           `gorm:"column:is_superadmin;default:false" json:"isSuperAdmin"`
	IsActive     bool           `gorm:"default:true" json:"isActive"`
	CreatedBy    string         `json:"createdBy"`
	UpdatedBy    string         `json:"updatedBy"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Admins []Admin `gorm:"foreignKey:RoleID" json:"-"`
}

// Department represents the organizational unit (e.g., Engineering, Finance)
type Department struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	Status      string         `gorm:"default:'active'" json:"status"` // active, inactive
	CreatedBy   string         `json:"createdBy"`
	UpdatedBy   string         `json:"updatedBy"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Permissions assigned to this department
	Permissions   []SystemPermission `gorm:"many2many:department_permissions;constraint:OnDelete:CASCADE" json:"permissions"`
	PermissionIDs []string           `gorm:"-" json:"permissionIds,omitempty"`

	// Admins belonging to this department (Many-to-Many)
	Admins []Admin `gorm:"many2many:admin_departments;constraint:OnDelete:CASCADE" json:"-"`
}

// Admin represents the user account
type Admin struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email        string         `gorm:"unique;not null" json:"email"`
	Name         string         `gorm:"column:name" json:"name"`
	LastName     string         `gorm:"column:last_name" json:"lastName"`
	Username     string         `gorm:"column:username" json:"username"`
	PhoneNumber  string         `gorm:"column:phone" json:"phone"`
	Position     string         `gorm:"column:position" json:"position"`
	Status       string         `gorm:"default:'active'" json:"status"`
	Password     string         `gorm:"-" json:"password,omitempty"`
	PasswordHash string         `gorm:"column:password_hash" json:"-"`
	
	// Identity
	RoleID       *uuid.UUID     `gorm:"type:uuid;column:role_id" json:"roleId"`

	CreatedBy    string         `json:"createdBy"`
	UpdatedBy    string         `json:"updatedBy"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Role       *AdminRole    `gorm:"foreignKey:RoleID;constraint:OnDelete:SET NULL" json:"role"`
	Departments []Department `gorm:"many2many:admin_departments;constraint:OnDelete:CASCADE" json:"departments"`
	DepartmentIDs []string   `gorm:"-" json:"departmentIds,omitempty"`
	RefreshTokens []AdminRefreshToken `gorm:"foreignKey:AdminUserID" json:"-"`
}

type AdminRefreshToken struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Token       string    `gorm:"unique;not null" json:"token"`
	ExpiryTime  time.Time `json:"expiryTime"`
	AdminUserID uuid.UUID `gorm:"type:uuid;not null" json:"adminUserId"`
	CreatedBy   string    `json:"createdBy"`
	UpdatedBy   string    `json:"updatedBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SystemPermission struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	ParentID    *string   `json:"parentId"`
	NameTh      string    `json:"nameTh"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"createdBy"`
	UpdatedBy   string    `json:"updatedBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Interfaces
type AdminRepository interface {
	GetByEmail(email string) (*Admin, error)
	GetByID(id uuid.UUID) (*Admin, error)
	GetPaginated(query AdminQuery) (*PaginatedAdminResponse, error)
	Create(admin *Admin) error
	Update(admin *Admin) error
	Delete(id uuid.UUID) error
}

type AdminUseCase interface {
	GetAdmins(query AdminQuery) (*PaginatedAdminResponse, error)
	GetAdminByID(id uuid.UUID) (*Admin, error)
	CreateAdmin(admin *Admin) error
	UpdateAdmin(admin *Admin) error
	DeleteAdmin(id uuid.UUID) error
}

type DepartmentRepository interface {
	GetPaginated(query DepartmentQuery) (*PaginatedDepartmentResponse, error)
	GetByID(id uuid.UUID) (*Department, error)
	GetAll() ([]Department, error)
	Create(dept *Department) error
	Update(dept *Department) error
	Delete(id uuid.UUID) error
}

type DepartmentUseCase interface {
	GetDepartments(query DepartmentQuery) (*PaginatedDepartmentResponse, error)
	GetAllDepartments() ([]Department, error)
	GetDepartmentByID(id uuid.UUID) (*Department, error)
	CreateDepartment(dept *Department) error
	UpdateDepartment(dept *Department) error
	DeleteDepartment(id uuid.UUID) error
}

// Queries & Responses
type AdminQuery struct {
	PageNumber int
	PageSize   int
	Email      string
	Name       string
}

type PaginatedAdminResponse struct {
	PageNumber int     `json:"pageNumber"`
	TotalItems int64   `json:"totalItems"`
	TotalPages int     `json:"totalPages"`
	Items      []Admin `json:"items"`
}

type DepartmentQuery struct {
	PageNumber int    `query:"page"`
	PageSize   int    `query:"pageSize"`
	Name       string `query:"name"`
	Status     string `query:"status"`
}

type PaginatedDepartmentResponse struct {
	PageNumber int          `json:"pageNumber"`
	TotalItems int64        `json:"totalItems"`
	TotalPages int          `json:"totalPages"`
	Items      []Department `json:"items"`
}

type AdminRoleRepository interface {
	GetAll() ([]AdminRole, error)
	GetByID(id uuid.UUID) (*AdminRole, error)
}

type AdminRoleUseCase interface {
	GetAllRoles() ([]AdminRole, error)
}

type SystemPermissionRepository interface {
	GetAll() ([]SystemPermission, error)
}

type SystemPermissionUseCase interface {
	GetAllPermissions() ([]SystemPermission, error)
}
