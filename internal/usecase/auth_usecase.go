package usecase

import (
	"errors"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type authUseCase struct {
	adminRepo domain.AdminRepository
}

func NewAuthUseCase(adminRepo domain.AdminRepository) domain.AuthUseCase {
	return &authUseCase{adminRepo: adminRepo}
}

func (u *authUseCase) Login(email, password string) (string, domain.User, error) {
	admin, err := u.adminRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", domain.User{}, errors.New("user not found")
		}
		return "", domain.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return "", domain.User{}, errors.New("invalid credentials")
	}

	role := admin.Department.Name
	permissions := make([]string, len(admin.Department.Permissions))
	copy(permissions, admin.Department.Permissions)

	user := domain.User{
		ID:          admin.ID.String(),
		Email:       admin.Email,
		Name:        admin.Name,
		Role:        role,
		Permissions: permissions,
	}

	token, err := jwtutil.GenerateToken(user)
	if err != nil {
		return "", domain.User{}, err
	}

	return token, user, nil
}
