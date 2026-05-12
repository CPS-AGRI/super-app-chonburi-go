package usecase

import (
	"errors"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type authUseCase struct {
	adminRepo domain.AdminRepository
	rtRepo    domain.AdminRefreshTokenRepository
}

func NewAuthUseCase(adminRepo domain.AdminRepository, rtRepo domain.AdminRefreshTokenRepository) domain.AuthUseCase {
	return &authUseCase{adminRepo: adminRepo, rtRepo: rtRepo}
}

func (u *authUseCase) Login(email, password string) (string, string, domain.User, error) {
	admin, err := u.adminRepo.GetByEmail(email)
	if err != nil {
		return "", "", nil, errors.New("user not found")
	}

	if admin == nil {
		return "", "", nil, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return "", "", nil, errors.New("invalid credentials")
	}

	token, err := jwtutil.GenerateToken(admin)
	if err != nil {
		return "", "", nil, err
	}

	// Generate Refresh Token
	refreshToken := domain.NewUUID()
	rt := &domain.AdminRefreshToken{
		ID:          domain.NewUUID(),
		Token:       refreshToken,
		ExpiryTime:  time.Now().Add(7 * 24 * time.Hour), // 7 days
		AdminUserId: admin.ID,
		CreatedBy:   admin.ID,
		UpdatedBy:   admin.ID,
		CreatedDate: time.Now(),
		UpdatedDate: time.Now(),
	}
	
	if err := u.rtRepo.Create(rt); err != nil {
		return "", "", nil, err
	}

	return token, refreshToken, admin, nil
}

func (u *authUseCase) RefreshToken(token string) (string, string, domain.User, error) {
	rt, err := u.rtRepo.GetByToken(token)
	if err != nil {
		return "", "", nil, errors.New("invalid refresh token")
	}

	if time.Now().After(rt.ExpiryTime) {
		_ = u.rtRepo.DeleteByToken(token)
		return "", "", nil, errors.New("refresh token expired")
	}

	admin, err := u.adminRepo.GetByID(rt.AdminUserId)
	if err != nil {
		return "", "", nil, err
	}

	newAccessToken, err := jwtutil.GenerateToken(admin)
	if err != nil {
		return "", "", nil, err
	}

	return newAccessToken, token, admin, nil
}

func (u *authUseCase) Logout(token string) error {
	return u.rtRepo.DeleteByToken(token)
}

func (u *authUseCase) Me(id string) (*domain.Admin, []string, error) {
	admin, err := u.adminRepo.GetByID(id)
	if err != nil {
		return nil, nil, err
	}

	permissions := []string{}
	
	// 1. If superadmin, add system permissions
	if admin.Role != nil && admin.Role.Type == "superadmin" {
		permissions = append(permissions, "MANAGE_CITY", "MANAGE_ADMINS", "MANAGE_DEPARTMENTS", "VIEW_ALL_REPORTS")
	}

	// 2. Add permissions from assigned modules
	uniqueKeys := make(map[string]bool)
	for _, dept := range admin.Departments {
		for _, module := range dept.Modules {
			if module.Key != nil && *module.Key != "" {
				uniqueKeys[*module.Key] = true
			}

			// Explicitly grant ModuleComplaintCenter if this is the Center module
			if module.ID == "d01b2ce5-34a9-498b-bba0-b1b8360f1ea9" || 
			   module.NameTh == "ศูนย์ร้องทุกข์" || 
			   module.NameEn == "Complaint Center" {
				uniqueKeys["ModuleComplaintCenter"] = true
			}
		}
	}

	for key := range uniqueKeys {
		permissions = append(permissions, key)
	}

	return admin, permissions, nil
}
