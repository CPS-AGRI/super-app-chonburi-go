package usecase

import (
	"errors"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"
	"github.com/google/uuid"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type authUseCase struct {
	adminRepo domain.AdminRepository
	rtRepo    domain.RefreshTokenRepository
}

func NewAuthUseCase(adminRepo domain.AdminRepository, rtRepo domain.RefreshTokenRepository) domain.AuthUseCase {
	return &authUseCase{adminRepo: adminRepo, rtRepo: rtRepo}
}

func (u *authUseCase) Login(email, password string) (string, string, domain.User, error) {
	admin, err := u.adminRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", domain.User{}, errors.New("user not found")
		}
		return "", "", domain.User{}, err
	}

	if admin == nil {
		return "", "", domain.User{}, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return "", "", domain.User{}, errors.New("invalid credentials")
	}

	// Aggregate permissions
	permissionMap := make(map[string]bool)
	
	if admin.Role != nil && admin.Role.IsSuperAdmin {
		// 1. Super Admin gets SPECIAL permissions (Manage City, etc.)
		permissionMap["MANAGE_CITY"] = true
		permissionMap["MANAGE_ADMINS"] = true
		permissionMap["MANAGE_DEPARTMENTS"] = true
		permissionMap["VIEW_ALL_REPORTS"] = true
	} else {
		// 2. Regular Admins get permissions ONLY from their Departments
		for _, dept := range admin.Departments {
			for _, p := range dept.Permissions {
				permissionMap[p.ID] = true
			}
		}
	}

	// Convert map back to slice
	finalPermissions := make([]string, 0, len(permissionMap))
	for pID := range permissionMap {
		finalPermissions = append(finalPermissions, pID)
	}

	user := domain.User{
		ID:          admin.ID.String(),
		Email:       admin.Email,
		Name:        admin.Name,
		Role:        admin.Role.Name,
		Permissions: finalPermissions,
	}

	token, err := jwtutil.GenerateToken(user)
	if err != nil {
		return "", "", domain.User{}, err
	}

	// Generate Refresh Token
	refreshToken := uuid.New().String()
	rt := &domain.AdminRefreshToken{
		Token:       refreshToken,
		ExpiryTime:  time.Now().Add(7 * 24 * time.Hour), // 7 days
		AdminUserID: admin.ID,
		CreatedBy:   admin.Email,
	}
	
	if err := u.rtRepo.Create(rt); err != nil {
		return "", "", domain.User{}, err
	}

	return token, refreshToken, user, nil
}

func (u *authUseCase) RefreshToken(token string) (string, string, domain.User, error) {
	rt, err := u.rtRepo.GetByToken(token)
	if err != nil {
		return "", "", domain.User{}, errors.New("invalid refresh token")
	}

	if time.Now().After(rt.ExpiryTime) {
		_ = u.rtRepo.DeleteByToken(token)
		return "", "", domain.User{}, errors.New("refresh token expired")
	}

	admin, err := u.adminRepo.GetByID(rt.AdminUserID)
	if err != nil {
		return "", "", domain.User{}, err
	}

	// Re-aggregate permissions
	permissionMap := make(map[string]bool)
	if admin.Role != nil && admin.Role.IsSuperAdmin {
		permissionMap["MANAGE_CITY"] = true
		permissionMap["MANAGE_ADMINS"] = true
		permissionMap["MANAGE_DEPARTMENTS"] = true
		permissionMap["VIEW_ALL_REPORTS"] = true
	} else {
		for _, dept := range admin.Departments {
			for _, p := range dept.Permissions {
				permissionMap[p.ID] = true
			}
		}
	}
	finalPermissions := make([]string, 0, len(permissionMap))
	for pID := range permissionMap {
		finalPermissions = append(finalPermissions, pID)
	}

	user := domain.User{
		ID:          admin.ID.String(),
		Email:       admin.Email,
		Name:        admin.Name,
		Role:        admin.Role.Name,
		Permissions: finalPermissions,
	}

	newAccessToken, err := jwtutil.GenerateToken(user)
	if err != nil {
		return "", "", domain.User{}, err
	}

	return newAccessToken, token, user, nil
}
