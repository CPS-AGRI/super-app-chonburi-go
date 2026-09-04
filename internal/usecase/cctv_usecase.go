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

func (u *cctvUseCase) RecordViewLog(input domain.CreateCCTVLogInput, userID uuid.UUID, clientIP, userAgent string) error {
	// Guard 1: Minimum Duration Rule (Noise Reduction)
	if input.DurationSeconds < 3 {
		return nil // Ignore noise / accidental clicks (< 3 seconds)
	}

	// Resolve CCTV UUID safely
	var cctvUUID uuid.UUID
	var err error
	if input.CCTVID != "" {
		cctvUUID, err = uuid.Parse(input.CCTVID)
	}

	// Fallback lookup if not a direct UUID
	if err != nil || cctvUUID == uuid.Nil {
		cctvs, _ := u.repo.GetPaginated(domain.CCTVQuery{PageNumber: 1, PageSize: 1, Name: input.CCTVID})
		if cctvs != nil && len(cctvs.Items) > 0 {
			cctvUUID = cctvs.Items[0].ID
		} else {
			allCCTVs, _ := u.repo.GetPaginated(domain.CCTVQuery{PageNumber: 1, PageSize: 1})
			if allCCTVs != nil && len(allCCTVs.Items) > 0 {
				cctvUUID = allCCTVs.Items[0].ID
			}
		}
	}

	if cctvUUID == uuid.Nil {
		return errors.New("cctv device not found")
	}

	// Guard 2: Future timestamp validation
	now := time.Now()
	if input.StartedAt.IsZero() {
		input.StartedAt = now.Add(-time.Duration(input.DurationSeconds) * time.Second)
	} else if input.StartedAt.After(now.Add(10 * time.Second)) {
		input.StartedAt = now
	}

	// Guard 3: Duration Clamping & Max cap (8 hours = 28,800s)
	if input.EndedAt != nil {
		maxSeconds := int(input.EndedAt.Sub(input.StartedAt).Seconds()) + 5
		if input.DurationSeconds > maxSeconds && maxSeconds > 0 {
			input.DurationSeconds = maxSeconds
		}
	}
	if input.DurationSeconds > 28800 {
		input.DurationSeconds = 28800
	}

	device := input.DeviceType
	if device == "" {
		device = "Desktop"
	}

	policyType := input.PolicyType
	if policyType == "" {
		policyType = "AdminPolicy"
	}

	log := domain.CCTVViewLog{
		ID:              uuid.New(),
		CCTVID:          cctvUUID,
		UserID:          userID,
		PolicyType:      policyType,
		DeviceType:      device,
		StartedAt:       input.StartedAt,
		EndedAt:         input.EndedAt,
		DurationSeconds: input.DurationSeconds,
		IPAddress:       clientIP,
		UserAgent:       userAgent,
		CreatedAt:       time.Now(),
	}

	return u.repo.CreateViewLog(&log)
}

func (u *cctvUseCase) GetRecentLogs(query domain.CCTVLogQuery) (*domain.PaginatedCCTVRecentLogResponse, error) {
	if query.PageNumber <= 0 {
		query.PageNumber = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 15
	}
	return u.repo.GetRecentLogs(query)
}

func (u *cctvUseCase) GetUserSummaryLogs(query domain.CCTVLogQuery) (*domain.PaginatedCCTVUserSummaryResponse, error) {
	if query.PageNumber <= 0 {
		query.PageNumber = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 15
	}
	return u.repo.GetUserSummaryLogs(query)
}

