package usecase

import (
	"errors"
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
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

func (u *adminUseCase) GetAdminByID(id uuid.UUID) (*domain.Admin, error) {
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
	existingAdmin, _ := u.adminRepo.GetByEmail(admin.Email)
	if existingAdmin != nil {
		return errors.New("email already exists")
	}

	if admin.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin.PasswordHash = string(hash)
	} else if admin.PasswordHash == "" {
		return errors.New("password is required")
	}

	return u.adminRepo.Create(admin)
}

func (u *adminUseCase) UpdateAdmin(admin *domain.Admin) error {
	existingAdmin, err := u.adminRepo.GetByID(admin.ID)
	if err != nil {
		return errors.New("admin not found")
	}

	// Avoid email conflict
	if admin.Email != existingAdmin.Email {
		conflictAdmin, _ := u.adminRepo.GetByEmail(admin.Email)
		if conflictAdmin != nil {
			return errors.New("email already in use")
		}
	}

	existingAdmin.Email = admin.Email
	existingAdmin.Name = admin.Name
	existingAdmin.LastName = admin.LastName
	existingAdmin.Username = admin.Username
	existingAdmin.PhoneNumber = admin.PhoneNumber
	existingAdmin.Position = admin.Position
	existingAdmin.Status = admin.Status
	existingAdmin.DepartmentID = admin.DepartmentID
	existingAdmin.Permissions = admin.Permissions

	if admin.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		existingAdmin.PasswordHash = string(hash)
	} else if admin.PasswordHash != "" {
        existingAdmin.PasswordHash = admin.PasswordHash
	}

	return u.adminRepo.Update(existingAdmin)
}

func (u *adminUseCase) DeleteAdmin(id uuid.UUID) error {
	return u.adminRepo.Delete(id)
}
