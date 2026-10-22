package usecase

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/internal/repository"
)

var thaiMonths = [...]string{
	"", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.",
	"ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค.",
}

type WaterLevelUseCase interface {
	GetStations(ctx context.Context) ([]domain.WaterStation, error)
	GetDeviceDetail(ctx context.Context, serial string) (*domain.WaterDeviceDetail, error)
	GetDeviceHistory(ctx context.Context, serial string, timeframe int) ([]domain.WaterHistoryPoint, error)
	GetSnapshots(ctx context.Context, serial, date string) ([]domain.WaterSnapshotLog, error)
	InitStream(ctx context.Context, serial string) (*domain.WaterStreamResult, error)
	TerminateStream(ctx context.Context, serial string) error
	GetActiveAlerts(ctx context.Context) ([]domain.WaterAlert, error)
	RelayWebRTCOffer(ctx context.Context, serial, sdpOffer string) (string, error)
}

type waterLevelUseCase struct {
	repo repository.WaterLevelRepository
}

func NewWaterLevelUseCase(repo repository.WaterLevelRepository) WaterLevelUseCase {
	return &waterLevelUseCase{
		repo: repo,
	}
}

func parseCoordinate(val any) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

func parseTimeRobust(isoStr string) (time.Time, error) {
	if isoStr == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05-07",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05+07:00",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, isoStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time format: %s", isoStr)
}

func formatThaiDateTime(isoStr string) (thaiTime string, iso8601 string) {
	if isoStr == "" {
		return "-", ""
	}
	t, err := parseTimeRobust(isoStr)
	if err != nil {
		return isoStr, isoStr
	}

	loc, err := time.LoadLocation("Asia/Bangkok")
	if err == nil {
		t = t.In(loc)
	}

	monthStr := ""
	if t.Month() >= 1 && int(t.Month()) < len(thaiMonths) {
		monthStr = thaiMonths[t.Month()]
	}

	thaiTime = fmt.Sprintf("%d %s %d, %02d:%02d น.", t.Day(), monthStr, t.Year(), t.Hour(), t.Minute())
	iso8601 = t.Format(time.RFC3339)
	return thaiTime, iso8601
}

func mapOnlineStatus(lastSeen string, threshold time.Duration) string {
	if lastSeen == "" {
		return "offline"
	}
	t, err := parseTimeRobust(lastSeen)
	if err != nil {
		return "offline"
	}
	if time.Since(t) <= threshold && time.Since(t) >= -5*time.Minute {
		return "online"
	}
	return "offline"
}

func mapWarningStatus(status string) string {
	switch status {
	case "WARNING":
		return "warning"
	case "CRITICAL":
		return "critical"
	case "EMERGENCY":
		return "emergency"
	default:
		return "normal"
	}
}

func (u *waterLevelUseCase) GetStations(ctx context.Context) ([]domain.WaterStation, error) {
	devices, err := u.repo.FetchGatewayDevices(ctx)
	if err != nil {
		log.Printf("⚠️ [WaterLevelUseCase] Error fetching devices: %v", err)
		return []domain.WaterStation{}, err
	}

	result := make([]domain.WaterStation, 0, len(devices))
	for _, dev := range devices {
		lat := parseCoordinate(dev.InstallLatitude)
		if lat == 0 {
			lat = dev.Latitude
		}
		lng := parseCoordinate(dev.InstallLongitude)
		if lng == 0 {
			lng = dev.Longitude
		}

		waterLevelM := dev.NetWaterLevelCM / 100.0
		waterLevelMaxM := dev.BottomToBankCM / 100.0

		pctToBank := dev.WaterLevelPct
		if pctToBank == 0 && dev.BottomToBankCM > 0 {
			pctToBank = math.Round((dev.NetWaterLevelCM/dev.BottomToBankCM)*1000) / 10
		}

		thaiTime, iso8601 := formatThaiDateTime(dev.LastSeen)
		onlineStatus := mapOnlineStatus(dev.LastSeen, 60*time.Minute)

		st := domain.WaterStation{
			ID:             dev.Serial,
			Serial:         dev.Serial,
			Name:           dev.Name,
			LocationName:   dev.Location,
			Lat:            lat,
			Lng:            lng,
			WaterLevel:     math.Round(waterLevelM*100) / 100,
			WaterLevelMax:  math.Round(waterLevelMaxM*100) / 100,
			RainAmount:     dev.RainAmount,
			WindSpeed:      dev.WindSpeed,
			WindDirection:  dev.WindDirection,
			Status:         mapWarningStatus(dev.WarningStatus),
			OnlineStatus:   onlineStatus,
			LastUpdated:    thaiTime,
			UpdatedAt:      iso8601,
			SnapshotURL:    dev.SnapshotURL0,
			CCTV:           dev.CCTVStreamURL,
			CommonSizeName: dev.Name,
			PctToBank:      pctToBank,
		}

		if st.SnapshotURL == "" {
			st.SnapshotURL = dev.CCTVPath0
		}

		result = append(result, st)
	}

	return result, nil
}

func (u *waterLevelUseCase) GetDeviceDetail(ctx context.Context, serial string) (*domain.WaterDeviceDetail, error) {
	envelope, err := u.repo.FetchGatewayDeviceDetail(ctx, serial)
	if err != nil {
		return nil, err
	}

	dev := envelope.Device
	cur := envelope.Current
	cfg := envelope.Config

	lat := parseCoordinate(dev.InstallLatitude)
	lng := parseCoordinate(dev.InstallLongitude)

	waterLevelM := cur.NetWaterLevelCM / 100.0
	waterLevelMaxM := cfg.BottomToBankCM / 100.0

	pctToBank := cur.WaterLevelPct
	if pctToBank == 0 && cfg.BottomToBankCM > 0 {
		pctToBank = math.Round((cur.NetWaterLevelCM/cfg.BottomToBankCM)*1000) / 10
	}

	snapURL := dev.SnapshotURL0
	if snapURL == "" {
		snapURL = dev.CCTVPath0
	}

	cctvURL := dev.CCTVStreamURL
	if cctvURL == "" {
		cctvURL = dev.SnapshotURL1
	}

	thaiTime, iso8601 := formatThaiDateTime(cur.Time)
	onlineStatus := mapOnlineStatus(cur.Time, 60*time.Minute)

	return &domain.WaterDeviceDetail{
		ID:            dev.Serial,
		Serial:        dev.Serial,
		Name:          dev.Name,
		Location:      dev.OrganizationName,
		Lat:           lat,
		Lng:           lng,
		Status:        mapWarningStatus(dev.WarningStatus),
		OnlineStatus:  onlineStatus,
		WaterLevel:    math.Round(waterLevelM*100) / 100,
		WaterLevelMax: math.Round(waterLevelMaxM*100) / 100,
		LastUpdated:   thaiTime,
		UpdatedAt:     iso8601,
		Battery:       cur.Battery,
		Solar:         cur.Solar,
		CCTVStreamURL: cctvURL,
		SnapshotURL:   snapURL,
		CCTV:          cctvURL,
		PctToBank:     pctToBank,
	}, nil
}

func (u *waterLevelUseCase) GetDeviceHistory(ctx context.Context, serial string, timeframe int) ([]domain.WaterHistoryPoint, error) {
	rawPoints, err := u.repo.FetchGatewayDeviceHistory(ctx, serial, timeframe)
	if err != nil {
		return []domain.WaterHistoryPoint{}, err
	}

	loc, _ := time.LoadLocation("Asia/Bangkok")
	points := make([]domain.WaterHistoryPoint, 0, len(rawPoints))

	for _, pt := range rawPoints {
		t, err := parseTimeRobust(pt.Time)
		if err != nil {
			t = time.Now()
		}
		if loc != nil {
			t = t.In(loc)
		}

		waterLevelM := pt.NetWaterLevelCM / 100.0

		points = append(points, domain.WaterHistoryPoint{
			Time:       pt.Time,
			WaterLevel: math.Round(waterLevelM*100) / 100,
			RainAmount: pt.RainAmount,
			WindSpeed:  pt.WindSpeed,
			HourLabel:  fmt.Sprintf("%02d:00", t.Hour()),
			DateLabel:  t.Format("02/01"),
			ISODate:    t.Format(time.RFC3339),
			Pct:        pt.WaterLevelPct,
		})
	}

	return points, nil
}

func (u *waterLevelUseCase) GetSnapshots(ctx context.Context, serial, date string) ([]domain.WaterSnapshotLog, error) {
	rawLogs, err := u.repo.FetchGatewaySnapshots(ctx, serial, date)
	if err != nil {
		return []domain.WaterSnapshotLog{}, err
	}

	logs := make([]domain.WaterSnapshotLog, 0, len(rawLogs))
	for _, l := range rawLogs {
		waterLevelM := l.NetWaterLevelCM / 100.0
		thaiTime, _ := formatThaiDateTime(l.Time)
		logs = append(logs, domain.WaterSnapshotLog{
			Time:       thaiTime,
			ImageURL:   l.ImageURL,
			WaterLevel: math.Round(waterLevelM*100) / 100,
		})
	}
	return logs, nil
}

func (u *waterLevelUseCase) InitStream(ctx context.Context, serial string) (*domain.WaterStreamResult, error) {
	return u.repo.InitGatewayStream(ctx, serial)
}

func (u *waterLevelUseCase) TerminateStream(ctx context.Context, serial string) error {
	return u.repo.TerminateGatewayStream(ctx, serial)
}

func (u *waterLevelUseCase) GetActiveAlerts(ctx context.Context) ([]domain.WaterAlert, error) {
	items, err := u.repo.FetchActiveAlerts(ctx)
	if err != nil {
		return []domain.WaterAlert{}, nil
	}

	alerts := make([]domain.WaterAlert, 0, len(items))
	for _, item := range items {
		thaiTime, _ := formatThaiDateTime(item.Time)
		alerts = append(alerts, domain.WaterAlert{
			ID:       fmt.Sprintf("%v", item.ID),
			Title:    fmt.Sprintf("แจ้งเตือนสถานี %s", item.Serial),
			Detail:   item.Message,
			Location: item.Serial,
			Time:     thaiTime,
			Severity: mapWarningStatus(item.AlertLevel),
			Unread:   true,
		})
	}
	return alerts, nil
}

func (u *waterLevelUseCase) RelayWebRTCOffer(ctx context.Context, serial, sdpOffer string) (string, error) {
	return u.repo.RelayWebRTCOffer(ctx, serial, sdpOffer)
}
