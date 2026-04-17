package usecase

import (
	"errors"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/internal/repository"
	"super-app-chonburi-go/pkg/jwtutil"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthUseCase interface {
	Login(req domain.LoginRequest) (*domain.AuthResponse, error)
}

type authUseCase struct {
	adminRepo repository.AdminRepository
}

func NewAuthUseCase(adminRepo repository.AdminRepository) AuthUseCase {
	return &authUseCase{adminRepo: adminRepo}
}

func (u *authUseCase) Login(req domain.LoginRequest) (*domain.AuthResponse, error) {
	admin, err := u.adminRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	role := "employee"
	switch admin.Department.Name {
	case "Super Administration":
		role = "superadmin"
	case "Supervisor Team":
		role = "supervisor"
	}

	user := domain.User{
		ID:    int(admin.ID),
		Email: admin.Email,
		Name:  admin.Name,
		Role:  role,
	}

	token, err := jwtutil.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}
