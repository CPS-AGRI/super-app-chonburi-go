package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"super-app-chonburi-go/config"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/internal/repository"
	"super-app-chonburi-go/internal/usecase"
)

type mockWeatherRepo struct {
	pattayaList []domain.PattayaStationMeta
	pattayaMap  map[string]domain.PattayaStationMeta
	sensors     []domain.WeatherSensor
	frames      []domain.WeatherTileFrame
}

func newMockWeatherRepo() *mockWeatherRepo {
	repo := &mockWeatherRepo{
		pattayaMap: make(map[string]domain.PattayaStationMeta),
	}
	// Add mock 165 stations
	for i := 1; i <= 165; i++ {
		serial := "S1-" + string(rune('A'+(i%26))) + string(rune('0'+(i%10)))
		meta := domain.PattayaStationMeta{
			Order:       i,
			PoleCode:    "FW001",
			Serial:      serial,
			Subdistrict: "หนองปรือ",
			District:    "บางละมุง",
			AdminZone:   "เมืองพัทยา",
			ActualLat:   12.92,
			ActualLng:   100.88,
		}
		repo.pattayaList = append(repo.pattayaList, meta)
		repo.pattayaMap[serial] = meta
	}

	// Add 3 live stations (2 matching 165, 1 other province station)
	repo.sensors = []domain.WeatherSensor{
		{
			Serial:        repo.pattayaList[0].Serial,
			Name:          "สถานี 1",
			Latitude:      12.924,
			Longitude:     100.882,
			Temp:          30.5,
			RainRate:      12.0,
			PM25:          22.0,
			WindDirection: 90.0,
			OnlineStatus:  "online",
			Status:        "normal",
		},
		{
			Serial:        repo.pattayaList[1].Serial,
			Name:          "สถานี 2",
			Latitude:      12.930,
			Longitude:     100.890,
			Temp:          31.0,
			RainRate:      60.0, // danger
			PM25:          40.0, // warning
			WindDirection: 180.0,
			OnlineStatus:  "online",
			Status:        "danger",
		},
		{
			Serial:            "S1-BANGSAEN-999",
			Name:              "หาดบางแสน",
			Latitude:          13.28,
			Longitude:         100.91,
			Temp:              29.0,
			RainRate:          0.0,
			PM25:              15.0,
			WindDirection:     0.0,
			WindDirectionText: "เหนือ (N)",
			OnlineStatus:      "online",
			Status:            "normal",
		},
	}

	return repo
}

func (m *mockWeatherRepo) GetActiveAccounts(ctx context.Context) ([]domain.WeatherAccount, error) {
	return []domain.WeatherAccount{
		{ID: uuid.New(), AccountName: "Chonburi Test Account", IsActive: true},
	}, nil
}

func (m *mockWeatherRepo) FetchAllSensors(ctx context.Context) ([]domain.WeatherSensor, error) {
	return m.sensors, nil
}

func (m *mockWeatherRepo) FetchMapFrames(ctx context.Context, variable, animType, date string, limit int) ([]domain.WeatherTileFrame, error) {
	return []domain.WeatherTileFrame{
		{
			TimeLabel:   "09:00 น.",
			DisplayTime: "09:00 น.",
			Formatted:   "202610210900",
			URL:         "https://s3.mueangsmart.com/fahfon-map-tiles/static/composites/rr/202610210900.webp",
		},
	}, nil
}

func (m *mockWeatherRepo) GetPattaya165List() []domain.PattayaStationMeta {
	return m.pattayaList
}

func (m *mockWeatherRepo) GetPattaya165Meta(serial string) (domain.PattayaStationMeta, bool) {
	meta, exists := m.pattayaMap[serial]
	return meta, exists
}

func (m *mockWeatherRepo) FlushCache(ctx context.Context) {}

func TestWeatherUseCase_PattayaScope(t *testing.T) {
	mockRepo := newMockWeatherRepo()
	uc := usecase.NewWeatherUseCase(mockRepo)

	ctx := context.Background()
	// Test scope=pattaya
	sensors, err := uc.GetSensors(ctx, domain.WeatherQueryFilter{Scope: "pattaya"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sensors) != 165 {
		t.Errorf("expected 165 stations for pattaya scope, got %d", len(sensors))
	}

	// First station should be online with temp 30.5
	if sensors[0].OnlineStatus != "online" || sensors[0].Temp != 30.5 {
		t.Errorf("expected station 0 to be online with temp 30.5, got onlineStatus=%s, temp=%f",
			sensors[0].OnlineStatus, sensors[0].Temp)
	}

	// Second station has rain 60 -> status should be danger
	if sensors[1].Status != "danger" {
		t.Errorf("expected station 1 status to be danger, got %s", sensors[1].Status)
	}

	// Station 2 should be offline fallback
	if sensors[2].OnlineStatus != "offline" {
		t.Errorf("expected station 2 to be offline, got %s", sensors[2].OnlineStatus)
	}
}

func TestWeatherUseCase_AllScope(t *testing.T) {
	mockRepo := newMockWeatherRepo()
	uc := usecase.NewWeatherUseCase(mockRepo)

	ctx := context.Background()
	// Test scope=all (Web Admin)
	sensors, err := uc.GetSensors(ctx, domain.WeatherQueryFilter{Scope: "all"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return all 3 live stations
	if len(sensors) != 3 {
		t.Errorf("expected 3 stations for all scope, got %d", len(sensors))
	}

	// Check Bangsaen station is present
	foundBangsaen := false
	for _, s := range sensors {
		if s.Serial == "S1-BANGSAEN-999" {
			foundBangsaen = true
			if s.WindDirectionText != "เหนือ (N)" {
				t.Errorf("expected wind direction text 'เหนือ (N)', got %s", s.WindDirectionText)
			}
		}
	}
	if !foundBangsaen {
		t.Errorf("expected Bangsaen station in all scope")
	}
}

func TestWeatherUseCase_RadarFrames(t *testing.T) {
	mockRepo := newMockWeatherRepo()
	uc := usecase.NewWeatherUseCase(mockRepo)

	ctx := context.Background()
	frames, err := uc.GetMapFrames(ctx, "rain", "forecast", "2026-10-21")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(frames) != 1 {
		t.Errorf("expected 1 frame, got %d", len(frames))
	}

	if frames[0].DisplayTime != "09:00 น." {
		t.Errorf("expected display time 09:00 น., got %s", frames[0].DisplayTime)
	}
}

func TestWeatherRepository_RealEmbeddedJSON(t *testing.T) {
	cfg := &config.Config{
		FahfonWeatherBaseURL: "http://10.1.2.25:50059",
	}
	repo := repository.NewWeatherRepository(nil, cfg, nil)

	pattayaList := repo.GetPattaya165List()
	if len(pattayaList) != 165 {
		t.Fatalf("expected real embedded pattaya list to have 165 stations, got %d", len(pattayaList))
	}

	first := pattayaList[0]
	if first.Order != 1 || first.PoleCode != "FW001001" || first.Serial != "S1-0000899472" {
		t.Errorf("unexpected first station: %+v", first)
	}

	meta, found := repo.GetPattaya165Meta("S1-0000899472")
	if !found || meta.PoleCode != "FW001001" {
		t.Errorf("failed to lookup station by serial: found=%v, meta=%+v", found, meta)
	}
}
