package domain

import (
	"time"

	"github.com/google/uuid"
)

type User interface {
	GetID() string
	GetEmail() string
	GetRole() string
}

type AdminRole struct {
	ID                 string    `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	NameTh             string    `gorm:"not null;column:name_th;index" json:"name_th"`
	NameEn             string    `gorm:"not null;default:'';column:name_en" json:"name_en"`
	Type               string    `gorm:"not null;column:type;index" json:"type"`
	Policy             string    `gorm:"not null;default:'';column:policy" json:"policy"`
	CanRegisterEnabled bool      `gorm:"not null;default:false;column:can_register_enabled" json:"can_register_enabled"`
	CanAssignModule    bool      `gorm:"not null;default:false;column:can_assign_module" json:"can_assign_module"`
	CreatedBy          string    `gorm:"not null;default:'';column:created_by" json:"created_by"`
	CreatedDate        time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:created_date" json:"created_date"`
	UpdatedBy          string    `gorm:"not null;default:'';column:updated_by" json:"updated_by"`
	UpdatedDate        time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:updated_date" json:"updated_date"`
}

func (AdminRole) TableName() string { return "admin_roles" }

type Module struct {
	ID                                 string    `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	NameTh                             string    `gorm:"not null;default:'';column:name_th" json:"name_th"`
	NameEn                             string    `gorm:"not null;default:'';column:name_en" json:"name_en"`
	CreatedBy                          string    `gorm:"not null;default:'';column:created_by" json:"created_by"`
	CreatedDate                        time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:created_date" json:"created_date"`
	UpdatedBy                          string    `gorm:"not null;default:'';column:updated_by" json:"updated_by"`
	UpdatedDate                        time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:updated_date" json:"updated_date"`
	IsUsedForUserRegistrationOnly      bool      `gorm:"not null;default:false;column:is_used_for_user_registration_only" json:"is_used_for_user_registration_only"`
	CanBeSelectedWithAdminUserSettings bool      `gorm:"not null;default:false;column:can_be_selected_with_admin_user_settings" json:"can_be_selected_with_admin_user_settings"`
	IsAdminOnly                        bool      `gorm:"not null;default:false;column:is_admin_only" json:"is_admin_only"`
	Sequence                           *int      `gorm:"column:sequence" json:"sequence"`
	Key                                *string   `gorm:"column:key" json:"key"`
	DashboardNameEn                    *string   `gorm:"column:dashboard_name_en" json:"dashboard_name_en"`
	DashboardNameTh                    *string   `gorm:"column:dashboard_name_th" json:"dashboard_name_th"`
	IsDashboard                        bool      `gorm:"not null;default:false;column:is_dashboard" json:"is_dashboard"`

	ModuleTypes []ModuleType `gorm:"foreignKey:ModuleId" json:"module_types,omitempty"`
}

func (Module) TableName() string { return "modules" }

type ModuleType struct {
	ID                                 string    `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	ModuleId                           string    `gorm:"type:uuid;not null;column:module_id" json:"module_id"`
	NameTh                             string    `gorm:"not null;default:'';column:name_th" json:"name_th"`
	NameEn                             string    `gorm:"not null;default:'';column:name_en" json:"name_en"`
	CreatedBy                          string    `gorm:"not null;default:'';column:created_by" json:"created_by"`
	CreatedDate                        time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:created_date" json:"created_date"`
	UpdatedBy                          string    `gorm:"not null;default:'';column:updated_by" json:"updated_by"`
	UpdatedDate                        time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:updated_date" json:"updated_date"`
	CanBeSelectedWithAdminUserSettings bool      `gorm:"not null;default:false;column:can_be_selected_with_admin_user_settings" json:"can_be_selected_with_admin_user_settings"`

	Module *Module `gorm:"foreignKey:ModuleId" json:"module,omitempty"`
}

func (ModuleType) TableName() string { return "module_types" }

type Department struct {
	ID            string    `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	Name          string    `gorm:"not null;column:name;index" json:"name"`
	Status        string    `gorm:"not null;column:status;index" json:"status"`
	Description   *string   `gorm:"column:description" json:"description"`
	CreatedBy     string    `gorm:"not null;default:'';column:created_by" json:"created_by"`
	CreatedDate   time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:created_date" json:"created_date"`
	UpdatedBy     string    `gorm:"not null;default:'';column:updated_by" json:"updated_by"`
	UpdatedDate   time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:updated_date" json:"updated_date"`
	Modules       []Module  `gorm:"many2many:department_modules;" json:"modules,omitempty"`
	ModuleTypeIds []string  `gorm:"-" json:"module_type_ids,omitempty"`
	ModuleIds     []string  `gorm:"-" json:"module_ids,omitempty"`
}

func (Department) TableName() string { return "departments" }

type DepartmentModule struct {
	ID           string    `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	DepartmentId string    `gorm:"type:uuid;not null;column:department_id;index" json:"department_id"`
	ModuleId     string    `gorm:"type:uuid;not null;column:module_id;index" json:"module_id"`
	CreatedBy    string    `gorm:"not null;default:'';column:created_by" json:"created_by"`
	CreatedDate  time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:created_date" json:"created_date"`
	UpdatedBy    string    `gorm:"not null;default:'';column:updated_by" json:"updated_by"`
	UpdatedDate  time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:updated_date" json:"updated_date"`
}

func (DepartmentModule) TableName() string { return "department_modules" }

type DepartmentModuleModuleType struct {
	DepartmentModuleId string `gorm:"type:uuid;primaryKey;column:department_module_id;index" json:"department_module_id"`
	ModuleTypeId       string `gorm:"type:uuid;primaryKey;column:module_type_id;index" json:"module_type_id"`
}

func (DepartmentModuleModuleType) TableName() string { return "department_module_module_types" }

type Admin struct {
	ID                        string    `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	Name                      string    `gorm:"not null;column:name;index" json:"name"`
	LastName                  string    `gorm:"not null;column:last_name;index" json:"last_name"`
	Email                     string    `gorm:"not null;column:email;unique;index" json:"email"`
	Phone                     string    `gorm:"not null;column:phone" json:"phone"`
	Position                  string    `gorm:"not null;column:position" json:"position"`
	PasswordHash              string    `gorm:"not null;column:password_hash" json:"-"`
	Password                  string    `gorm:"-" json:"password,omitempty"`
	RoleId                    *string   `gorm:"type:uuid;column:role_id;index" json:"role_id"`
	CreatedBy                 string    `gorm:"not null;default:'';column:created_by" json:"created_by"`
	CreatedDate               time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:created_date" json:"created_date"`
	UpdatedBy                 string    `gorm:"not null;default:'';column:updated_by" json:"updated_by"`
	UpdatedDate               time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:updated_date" json:"updated_date"`
	VerifyForgotPasswordToken *string   `gorm:"column:verify_forgot_password_token" json:"verify_forgot_password_token"`
	VerifyRegistrationToken   string    `gorm:"not null;default:'';column:verify_registration_token" json:"verify_registration_token"`

	Role          *AdminRole   `gorm:"foreignKey:RoleId" json:"role,omitempty"`
	Departments   []Department `gorm:"many2many:admin_departments;" json:"departments"`
	DepartmentIds []string     `gorm:"-" json:"department_ids,omitempty"`
}

func (a Admin) GetID() string    { return a.ID }
func (a Admin) GetEmail() string { return a.Email }
func (a Admin) GetRole() string {
	if a.Role != nil {
		return a.Role.Type
	}
	return ""
}
func (Admin) TableName() string { return "admin_users" }

type AdminRefreshToken struct {
	ID          string    `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	Token       string    `gorm:"unique;not null;column:token;index" json:"token"`
	ExpiryTime  time.Time `gorm:"not null;type:timestamptz;column:expiry_time;index" json:"expiry_time"`
	AdminUserId string    `gorm:"type:uuid;not null;column:admin_user_id;index" json:"admin_user_id"`
	CreatedBy   string    `gorm:"not null;default:'';column:created_by" json:"created_by"`
	CreatedDate time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:created_date" json:"created_date"`
	UpdatedBy   string    `gorm:"not null;default:'';column:updated_by" json:"updated_by"`
	UpdatedDate time.Time `gorm:"not null;type:timestamptz;default:'-infinity';column:updated_date" json:"updated_date"`
}

func (AdminRefreshToken) TableName() string { return "admin_refresh_tokens" }

type AuditLog struct {
	ID                 string    `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	TraceId            string    `gorm:"not null;column:trace_id;index" json:"trace_id"`
	UserId             string    `gorm:"not null;column:user_id;index" json:"user_id"`
	RoleId             string    `gorm:"not null;column:role_id;index" json:"role_id"`
	MunicipalityIds    string    `gorm:"not null;column:municipality_ids" json:"municipality_ids"`
	Method             string    `gorm:"not null;column:method" json:"method"`
	Path               string    `gorm:"not null;column:path" json:"path"`
	ResponseStatusCode int       `gorm:"not null;column:response_status_code" json:"response_status_code"`
	RequestTime        time.Time `gorm:"not null;type:timestamptz;column:request_time" json:"request_time"`
	ResponseTime       time.Time `gorm:"not null;type:timestamptz;column:response_time" json:"response_time"`
}

func (AuditLog) TableName() string { return "audit_logs" }

func NewUUID() string {
	return uuid.New().String()
}

type AdminRepository interface {
	GetByEmail(email string) (*Admin, error)
	GetByID(id string) (*Admin, error)
	GetByVerifyRegistrationToken(token string) (*Admin, error)
	GetByVerifyForgotPasswordToken(token string) (*Admin, error)
	GetPaginated(query AdminQuery) (*PaginatedAdminResponse, error)
	Create(admin *Admin) error
	Update(admin *Admin) error
	UpdateFields(id string, fields map[string]interface{}) error
	Delete(id string) error
	GetAllModuleKeys() ([]string, error)
}

type AdminUseCase interface {
	GetAdmins(query AdminQuery) (*PaginatedAdminResponse, error)
	GetAdminByID(id string) (*Admin, error)
	CreateAdmin(admin *Admin) error
	UpdateAdmin(admin *Admin) error
	DeleteAdmin(id string) error
}

type AdminRoleRepository interface {
	GetAll() ([]AdminRole, error)
	GetByID(id string) (*AdminRole, error)
	Create(role *AdminRole) error
	Update(role *AdminRole) error
	Delete(id string) error
}

type AdminRoleUseCase interface {
	GetAllRoles() ([]AdminRole, error)
	GetRoleByID(id string) (*AdminRole, error)
	CreateRole(role *AdminRole) error
	UpdateRole(role *AdminRole) error
	DeleteRole(id string) error
}

type DepartmentRepository interface {
	GetPaginated(query DepartmentQuery) (*PaginatedDepartmentResponse, error)
	GetAll() ([]Department, error)
	GetByID(id string) (*Department, error)
	Create(dept *Department) error
	Update(dept *Department) error
	Delete(id string) error
}

type DepartmentUseCase interface {
	GetDepartments(query DepartmentQuery) (*PaginatedDepartmentResponse, error)
	GetAllDepartments() ([]Department, error)
	GetDepartmentByID(id string) (*Department, error)
	CreateDepartment(dept *Department) error
	UpdateDepartment(dept *Department) error
	DeleteDepartment(id string) error
}

type ModuleRepository interface {
	GetAll() ([]Module, error)
	GetByDepartmentID(deptID string) ([]Module, error)
	AssignToDepartment(deptID string, moduleIDs []string) error
	Create(module *Module) error
	Update(module *Module) error
	Delete(id string) error
}

type ModuleUseCase interface {
	GetModulesForUser(adminID string) ([]Module, error)
	GetAllModules() ([]Module, error)
	AssignModulesToDepartment(deptID string, moduleIDs []string) error
	CreateModule(module *Module) error
	UpdateModule(module *Module) error
	DeleteModule(id string) error
}

type ModuleTypeRepository interface {
	GetAll() ([]ModuleType, error)
	GetByModuleID(moduleID string) ([]ModuleType, error)
	AssignToDepartmentModule(deptID, moduleID string, typeIDs []string) error
	Create(mt *ModuleType) error
	Update(mt *ModuleType) error
	Delete(id string) error
}

type ModuleTypeUseCase interface {
	GetAllTypes() ([]ModuleType, error)
	GetTypesByModule(moduleID string) ([]ModuleType, error)
	AssignTypesToDepartmentModule(deptID, moduleID string, typeIDs []string) error
	CreateType(mt *ModuleType) error
	UpdateType(mt *ModuleType) error
	DeleteType(id string) error
}

type AdminRefreshTokenRepository interface {
	Create(rt *AdminRefreshToken) error
	GetByToken(token string) (*AdminRefreshToken, error)
	DeleteByToken(token string) error
	DeleteByAdminID(adminID string) error
}

type AuthUseCase interface {
	Login(email, password string) (string, string, User, error)
	RefreshToken(token string) (string, string, User, error)
	Logout(token string) error
	Me(id string) (*Admin, []string, error)
	ForgotPassword(email string) error
	VerifyToken(token, tokenType string) (*Admin, error)
	ResetPassword(token, newPassword, tokenType string) error
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token     string `json:"token" validate:"required"`
	Password  string `json:"password" validate:"required,min=8"`
	TokenType string `json:"type"` // "activation" or "reset"
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AdminQuery struct {
	PageNumber int
	PageSize   int
	Name       string
	Email      string
}

type PaginatedAdminResponse struct {
	Items      []Admin `json:"items"`
	TotalItems int64   `json:"total_items"`
	PageNumber int     `json:"page_number"`
	TotalPages int     `json:"total_pages"`
}

type DepartmentQuery struct {
	PageNumber int
	PageSize   int
	Name       string
	Status     string
}

type PaginatedResponse[T any] struct {
	Items       T     `json:"items"`
	TotalItems  int64 `json:"total_items"`
	PageNumber  int   `json:"page_number"`
	PageSize    int   `json:"page_size"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

type PaginatedDepartmentResponse struct {
	Items      []Department `json:"items"`
	TotalItems int64        `json:"total_items"`
	PageNumber int          `json:"page_number"`
	TotalPages int          `json:"total_pages"`
}
