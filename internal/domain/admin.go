package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type AdminDepartment struct {
	ID          uuid.UUID           `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:id" json:"id"`
	Name        string         `gorm:"unique;column:name" json:"name"`
	Description *string        `gorm:"column:description" json:"description"`
	IsActive    bool           `gorm:"column:isActive;default:true" json:"isActive"`
	CreatedAt   time.Time      `gorm:"column:createdAt;default:now()" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"column:updatedAt" json:"updatedAt"`
	Permissions pq.StringArray `gorm:"type:text[];column:permissions" json:"permissions"`

	Admins []Admin `gorm:"foreignKey:DepartmentID" json:"-"`
}

func (AdminDepartment) TableName() string {
	return "AdminDepartment"
}

type Admin struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4();column:id" json:"id"`
	Email        string    `gorm:"unique;column:email" json:"email"`
	Name         string    `gorm:"column:name" json:"name"`
	LastName     string    `gorm:"column:lastName" json:"lastName"`
	Username     string    `gorm:"column:username" json:"username"`
	PhoneNumber  string    `gorm:"column:phoneNumber" json:"phone"`
	Position     string    `gorm:"column:position" json:"position"`
	Status       string    `gorm:"column:status;default:'active'" json:"status"`
	Password     string    `gorm:"-" json:"password,omitempty"`     // Transient field for request
	PasswordHash string    `gorm:"column:passwordHash" json:"-"`    // Ignored in JSON response
	DepartmentID uuid.UUID      `gorm:"type:uuid;column:departmentId" json:"departmentId"`
	Permissions  pq.StringArray `gorm:"type:text[];column:permissions" json:"permissions"`
	CreatedAt    time.Time `gorm:"column:createdAt;default:now()" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updatedAt" json:"updatedAt"`

	Department AdminDepartment `gorm:"foreignKey:DepartmentID;constraint:OnDelete:RESTRICT;" json:"department"`
}

func (Admin) TableName() string {
	return "Admin"
}

type AdminQuery struct {
	PageNumber   int
	PageSize     int
	Email        string
	Name         string
	DepartmentID string
}

type PaginatedAdminResponse struct {
	PageNumber int       `json:"pageNumber"`
	TotalItems int64     `json:"totalItems"`
	TotalPages int       `json:"totalPages"`
	Items      []Admin   `json:"items"`
}

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

type AdminDepartmentQuery struct {
	PageNumber int
	PageSize   int
	Name       string
}

type PaginatedAdminDepartmentResponse struct {
	PageNumber int               `json:"pageNumber"`
	TotalItems int64             `json:"totalItems"`
	TotalPages int               `json:"totalPages"`
	Items      []AdminDepartment `json:"items"`
}

type AdminDepartmentRepository interface {
	GetPaginated(query AdminDepartmentQuery) (*PaginatedAdminDepartmentResponse, error)
	GetByID(id uuid.UUID) (*AdminDepartment, error)
	Create(department *AdminDepartment) error
	Update(department *AdminDepartment) error
	Delete(id uuid.UUID) error
}

type AdminDepartmentUseCase interface {
	GetDepartments(query AdminDepartmentQuery) (*PaginatedAdminDepartmentResponse, error)
	GetDepartmentByID(id uuid.UUID) (*AdminDepartment, error)
	CreateDepartment(department *AdminDepartment) error
	UpdateDepartment(department *AdminDepartment) error
	DeleteDepartment(id uuid.UUID) error
}

type SystemPermission struct {
	ID          string    `gorm:"primaryKey;column:id" json:"id"`                      // e.g. "MANAGE_COMPLAINTS"
	NameTh      string    `gorm:"column:nameTh" json:"nameTh"`                         // e.g. "ระบบเรื่องร้องเรียน"
	Description string    `gorm:"column:description" json:"description"`               // module description
	CreatedAt   time.Time `gorm:"column:createdAt;default:now()" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updatedAt;default:now()" json:"updatedAt"`
}

func (SystemPermission) TableName() string {
	return "SystemPermission"
}

type SystemPermissionRepository interface {
	GetAll() ([]SystemPermission, error)
}

type SystemPermissionUseCase interface {
	GetAllPermissions() ([]SystemPermission, error)
}
