package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"testing"

	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type stubCCTVRepo struct {
	cam     *domain.CCTV
	updated map[string]interface{}
}

func (s *stubCCTVRepo) Create(*domain.CCTV) error                                         { return nil }
func (s *stubCCTVRepo) GetPaginated(domain.CCTVQuery) (*domain.PaginatedCCTVResponse, error) {
	return nil, nil
}
func (s *stubCCTVRepo) FindByID(uuid.UUID) (*domain.CCTV, error) { return s.cam, nil }
func (s *stubCCTVRepo) Update(_ uuid.UUID, updates map[string]interface{}) error {
	s.updated = updates
	return nil
}
func (s *stubCCTVRepo) Delete(uuid.UUID, string) error                       { return nil }
func (s *stubCCTVRepo) GetHealthTargets() ([]domain.CCTVHealthTarget, error) { return nil, nil }
func (s *stubCCTVRepo) BatchUpdateStatus([]domain.CCTVHealthUpdate) (int64, error) {
	return 0, nil
}
func (s *stubCCTVRepo) CreateViewLog(*domain.CCTVViewLog) error { return nil }
func (s *stubCCTVRepo) GetRecentLogs(domain.CCTVLogQuery) (*domain.PaginatedCCTVRecentLogResponse, error) {
	return nil, nil
}
func (s *stubCCTVRepo) GetUserSummaryLogs(domain.CCTVLogQuery) (*domain.PaginatedCCTVUserSummaryResponse, error) {
	return nil, nil
}

func TestCheckCameraHealth_ViaRemoteCheckerSigned(t *testing.T) {
	camID := uuid.New()
	secret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	repo := &stubCCTVRepo{
		cam: &domain.CCTV{
			ID:        camID,
			Name:      "Cam1",
			StreamURL: "https://cctv-request.chon.go.th:2001/webrtc/g2cam1/?token=abc",
			Status:    "OFFLINE",
		},
	}

	var gotSig, gotTS, gotNonce string
	var gotBody []byte
	uc := NewCCTVUseCaseWithChecker(repo, "http://10.1.2.25:50061", secret).(*cctvUseCase)
	uc.httpClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			gotSig = r.Header.Get("X-Signature")
			gotTS = r.Header.Get("X-Timestamp")
			gotNonce = r.Header.Get("X-Nonce")
			gotBody, _ = io.ReadAll(r.Body)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"success":true,"status":"ONLINE"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	status, err := uc.CheckCameraHealth(camID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if status != "ONLINE" || repo.updated["status"] != "ONLINE" {
		t.Fatalf("expected ONLINE update, got status=%s updated=%+v", status, repo.updated)
	}

	bodySum := sha256.Sum256(gotBody)
	canonical := strings.Join([]string{"POST", "/api/v1/cctv/probe", gotTS, gotNonce, hex.EncodeToString(bodySum[:])}, "\n")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical))
	expected := hex.EncodeToString(mac.Sum(nil))
	if gotSig != expected {
		t.Fatalf("signature mismatch: got=%s want=%s", gotSig, expected)
	}
}

