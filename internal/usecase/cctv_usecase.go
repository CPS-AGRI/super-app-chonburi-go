package usecase

import (
	"errors"
	"time"

	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
)

type cctvUseCase struct {
	repo domain.AdminCCTVRepository
}

// NewCCTVUseCase creates a new AdminCCTVUseCase.
func NewCCTVUseCase(repo domain.AdminCCTVRepository) domain.AdminCCTVUseCase {
	return &cctvUseCase{repo: repo}
}

func (u *cctvUseCase) CreateCCTV(cctv *domain.CCTV) error {
	if cctv.Name == "" {
		return errors.New("camera name is required")
	}
	if cctv.Latitude == 0 || cctv.Longitude == 0 {
		return errors.New("valid latitude and longitude coordinates are required")
	}
	if cctv.ID == uuid.Nil {
		cctv.ID = uuid.New()
	}
	cctv.CreatedAt = time.Now()
	cctv.UpdatedAt = time.Now()
	return u.repo.Create(cctv)
}

func (u *cctvUseCase) GetCCTVs(query domain.CCTVQuery) (*domain.PaginatedCCTVResponse, error) {
	if query.PageNumber <= 0 {
		query.PageNumber = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	return u.repo.GetPaginated(query)
}

func (u *cctvUseCase) GetCCTVRequests(query domain.CCTVRequestQuery) (*domain.PaginatedCCTVRequestResponse, error) {
	if query.PageNumber <= 0 {
		query.PageNumber = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	return u.repo.GetRequestsPaginated(query)
}

func (u *cctvUseCase) ApproveRequest(id uuid.UUID, responseFileURL string, approvedBy uuid.UUID) error {
	req, err := u.repo.GetRequestByID(id)
	if err != nil {
		return err
	}
	if req == nil {
		return errors.New("cctv request not found")
	}

	req.Status = "APPROVED"
	req.ResponseFileURL = &responseFileURL
	req.ApprovedByID = &approvedBy
	req.UpdatedAt = time.Now()

	return u.repo.UpdateRequest(req)
}

func (u *cctvUseCase) RejectRequest(id uuid.UUID, reason string, approvedBy uuid.UUID) error {
	req, err := u.repo.GetRequestByID(id)
	if err != nil {
		return err
	}
	if req == nil {
		return errors.New("cctv request not found")
	}

	req.Status = "REJECTED"
	req.RejectReason = &reason
	req.ApprovedByID = &approvedBy
	req.UpdatedAt = time.Now()

	return u.repo.UpdateRequest(req)
}

func (u *cctvUseCase) DeleteCCTV(id uuid.UUID, adminID string) error {
	return u.repo.Delete(id, adminID)
}
