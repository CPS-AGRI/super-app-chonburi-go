package usecase

import (
	"errors"
	"super-app-chonburi-go/internal/domain"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type adminUseCase struct {
	adminRepo domain.AdminRepository
}

func NewAdminUseCase(adminRepo domain.AdminRepository) domain.AdminUseCase {
	return &adminUseCase{adminRepo: adminRepo}
}

func (u *adminUseCase) GetAdmins(query domain.AdminQuery) (*domain.PaginatedAdminResponse, error) {
	if query.PageNumber < 1 {
		query.PageNumber = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	return u.adminRepo.GetPaginated(query)
}

func (u *adminUseCase) GetAdminByID(id string) (*domain.Admin, error) {
	admin, err := u.adminRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, errors.New("admin not found")
	}
	return admin, nil
}

func (u *adminUseCase) CreateAdmin(admin *domain.Admin) error {
	if admin.Email == "" {
		return errors.New("email is required")
	}
	if admin.Password == "" {
		return errors.New("password is required")
	}

	existingAdmin, _ := u.adminRepo.GetByEmail(admin.Email)
	if existingAdmin != nil {
		return errors.New("email already exists")
	}

	if admin.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin.PasswordHash = string(hashed)
	}

	admin.ID = domain.NewUUID()
	admin.CreatedBy = "system"
	admin.UpdatedBy = "system"
	admin.CreatedDate = time.Now()
	admin.UpdatedDate = time.Now()

	// Map Department IDs to objects
	admin.Departments = []domain.Department{}
	for _, deptID := range admin.DepartmentIds {
		admin.Departments = append(admin.Departments, domain.Department{ID: deptID})
	}

	return u.adminRepo.Create(admin)
}

func (u *adminUseCase) UpdateAdmin(admin *domain.Admin) error {
	existingAdmin, err := u.adminRepo.GetByID(admin.ID)
	if err != nil || existingAdmin == nil {
		return errors.New("admin not found")
	}

	existingAdmin.Email = admin.Email
	existingAdmin.Name = admin.Name
	existingAdmin.LastName = admin.LastName
	existingAdmin.Phone = admin.Phone
	existingAdmin.Position = admin.Position
	existingAdmin.RoleId = admin.RoleId
	existingAdmin.UpdatedBy = "system"
	existingAdmin.UpdatedDate = time.Now()

	if admin.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		existingAdmin.PasswordHash = string(hashed)
	}

	// Map Department IDs to objects
	existingAdmin.Departments = []domain.Department{}
	for _, deptID := range admin.DepartmentIds {
		existingAdmin.Departments = append(existingAdmin.Departments, domain.Department{ID: deptID})
	}

	return u.adminRepo.Update(existingAdmin)
}

func (u *adminUseCase) DeleteAdmin(id string) error {
	return u.adminRepo.Delete(id)
}
