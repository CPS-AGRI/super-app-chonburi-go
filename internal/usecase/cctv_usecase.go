package usecase

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
)

type cctvUseCase struct {
	repo          domain.AdminCCTVRepository
	checkerURL    string
	checkerSecret string
	httpClient    *http.Client
}

// NewCCTVUseCase creates a new AdminCCTVUseCase.
func NewCCTVUseCase(repo domain.AdminCCTVRepository) domain.AdminCCTVUseCase {
	return &cctvUseCase{
		repo:       repo,
		httpClient: &http.Client{Timeout: 8 * time.Second},
	}
}

// NewCCTVUseCaseWithChecker creates an AdminCCTVUseCase that offloads stream health checks
// to the dedicated chon-cctv-checker microservice over WireGuard when configured.
func NewCCTVUseCaseWithChecker(repo domain.AdminCCTVRepository, checkerURL, checkerSecret string) domain.AdminCCTVUseCase {
	return &cctvUseCase{
		repo:          repo,
		checkerURL:    strings.TrimRight(strings.TrimSpace(checkerURL), "/"),
		checkerSecret: strings.TrimSpace(checkerSecret),
		httpClient:    &http.Client{Timeout: 8 * time.Second},
	}
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
	if cctv.Status == "" {
		cctv.Status = "ONLINE"
	}
	cctv.CreatedAt = time.Now()
	cctv.UpdatedAt = time.Now()
	if err := u.repo.Create(cctv); err != nil {
		return err
	}

	// Verify stream health asynchronously
	if cctv.StreamURL != "" {
		go func(camID uuid.UUID, stream string) {
			actualStatus := u.probeStreamStatus(stream)
			_ = u.repo.Update(camID, map[string]interface{}{"status": actualStatus})
		}(cctv.ID, cctv.StreamURL)
	}

	return nil
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

func (u *cctvUseCase) GetCCTVByID(id uuid.UUID) (*domain.CCTV, error) {
	return u.repo.FindByID(id)
}

func (u *cctvUseCase) UpdateCCTV(id uuid.UUID, updates map[string]interface{}) (*domain.CCTV, error) {
	if streamURL, ok := updates["stream_url"].(string); ok && streamURL != "" {
		// When stream URL is modified, re-verify its health
		updates["status"] = u.probeStreamStatus(streamURL)
	}
	if err := u.repo.Update(id, updates); err != nil {
		return nil, err
	}
	return u.repo.FindByID(id)
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

	// Fallback lookup by Name if not a direct UUID
	if err != nil || cctvUUID == uuid.Nil {
		cctvs, _ := u.repo.GetPaginated(domain.CCTVQuery{PageNumber: 1, PageSize: 1, Name: input.CCTVID})
		if cctvs != nil && len(cctvs.Items) > 0 {
			cctvUUID = cctvs.Items[0].ID
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

func (u *cctvUseCase) CheckCameraHealth(id uuid.UUID) (string, error) {
	cam, err := u.repo.FindByID(id)
	if err != nil {
		return "OFFLINE", err
	}
	status := u.probeStreamStatus(cam.StreamURL)
	if strings.ToUpper(cam.Status) != status {
		_ = u.repo.Update(id, map[string]interface{}{"status": status})
	}
	return status, nil
}

func (u *cctvUseCase) GetHealthTargets() ([]domain.CCTVHealthTarget, error) {
	return u.repo.GetHealthTargets()
}

func (u *cctvUseCase) SyncCameraHealthStatuses(updates []domain.CCTVHealthUpdate) (int64, error) {
	return u.repo.BatchUpdateStatus(updates)
}

func (u *cctvUseCase) probeStreamStatus(streamURL string) string {
	if u.checkerURL != "" && len(u.checkerSecret) >= 32 {
		if status, ok := u.probeViaRemoteChecker(streamURL); ok {
			return status
		}
	}
	return CheckCameraStatus(streamURL)
}

func (u *cctvUseCase) probeViaRemoteChecker(streamURL string) (string, bool) {
	const probePath = "/api/v1/cctv/probe"
	payload, err := json.Marshal(map[string]string{"stream_url": strings.TrimSpace(streamURL)})
	if err != nil {
		return "OFFLINE", false
	}

	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "OFFLINE", false
	}
	nonce := hex.EncodeToString(nonceBytes)
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	bodySum := sha256.Sum256(payload)
	canonical := strings.Join([]string{
		http.MethodPost,
		probePath,
		ts,
		nonce,
		hex.EncodeToString(bodySum[:]),
	}, "\n")
	mac := hmac.New(sha256.New, []byte(u.checkerSecret))
	mac.Write([]byte(canonical))
	sig := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequest(http.MethodPost, u.checkerURL+probePath, bytes.NewReader(payload))
	if err != nil {
		return "OFFLINE", false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Timestamp", ts)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Signature", sig)

	client := u.httpClient
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("⚠️ [CCTV Probe] Remote checker unreachable: %v", err)
		return "OFFLINE", false
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("⚠️ [CCTV Probe] Remote checker returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
		return "OFFLINE", false
	}

	var parsed struct {
		Success bool   `json:"success"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil || !parsed.Success {
		return "OFFLINE", false
	}
	normalized := strings.ToUpper(strings.TrimSpace(parsed.Status))
	if normalized != "ONLINE" && normalized != "OFFLINE" {
		return "OFFLINE", false
	}
	return normalized, true
}


