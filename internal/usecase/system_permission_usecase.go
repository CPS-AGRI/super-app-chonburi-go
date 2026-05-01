package usecase

import (
	"errors"
	"super-app-chonburi-go/internal/domain"
)

type systemPermissionUseCase struct {
	repo domain.SystemPermissionRepository
}

func NewSystemPermissionUseCase(repo domain.SystemPermissionRepository) domain.SystemPermissionUseCase {
	return &systemPermissionUseCase{repo: repo}
}

func (u *systemPermissionUseCase) GetAllPermissions() ([]domain.SystemPermission, error) {
	return u.repo.GetAll()
}

func (u *systemPermissionUseCase) CreatePermission(p *domain.SystemPermission) error {
	if p.ParentID != nil && *p.ParentID != "" {
		// Check if parent exists
		all, _ := u.repo.GetAll()
		exists := false
		for _, v := range all {
			if v.ID == *p.ParentID {
				exists = true
				break
			}
		}
		if !exists {
			return errors.New("Parent permission not found")
		}
	}
	return u.repo.Create(p)
}

func (u *systemPermissionUseCase) UpdatePermission(p *domain.SystemPermission) error {
	if p.ParentID != nil && *p.ParentID != "" {
		// Check if parent exists
		all, _ := u.repo.GetAll()
		exists := false
		for _, v := range all {
			if v.ID == *p.ParentID {
				exists = true
				break
			}
		}
		if !exists {
			return errors.New("Parent permission not found")
		}
	}
	return u.repo.Update(p)
}

func (u *systemPermissionUseCase) DeletePermission(id string) error {
	// Check if this permission is a parent to others
	all, _ := u.repo.GetAll()
	for _, v := range all {
		if v.ParentID != nil && *v.ParentID == id {
			return errors.New("Cannot delete permission because it has child permissions")
		}
	}
	return u.repo.Delete(id)
}
