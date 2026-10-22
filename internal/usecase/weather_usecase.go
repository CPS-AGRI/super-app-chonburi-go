package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/internal/repository"
)

type WeatherUseCase interface {
	GetSensors(ctx context.Context, filter domain.WeatherQueryFilter) ([]domain.WeatherSensor, error)
	GetSensorBySerial(ctx context.Context, serial string) (*domain.WeatherSensor, error)
	GetDashboardStats(ctx context.Context) (*domain.WeatherDashboardStats, error)
	GetMapFrames(ctx context.Context, variable, animType, date string) ([]domain.WeatherTileFrame, error)
	FlushCache(ctx context.Context)
}

type weatherUseCase struct {
	repo repository.WeatherRepository
}

func NewWeatherUseCase(repo repository.WeatherRepository) WeatherUseCase {
	return &weatherUseCase{
		repo: repo,
	}
}

func (u *weatherUseCase) FlushCache(ctx context.Context) {
	u.repo.FlushCache(ctx)
}

func (u *weatherUseCase) GetSensors(ctx context.Context, filter domain.WeatherQueryFilter) ([]domain.WeatherSensor, error) {
	allSensors, err := u.repo.FetchAllSensors(ctx)
	if err != nil {
		return nil, err
	}

	// Scope: Pattaya (Exactly 165 Stations mapped to Official Excel Registry)
	if strings.ToLower(filter.Scope) == "pattaya" {
		liveMap := make(map[string]domain.WeatherSensor, len(allSensors))
		for _, s := range allSensors {
			liveMap[s.Serial] = s
		}

		pattayaList := u.repo.GetPattaya165List()
		nowBkk := time.Now().In(time.FixedZone("Asia/Bangkok", 7*3600))
		timeStr := nowBkk.Format("15:04 น.")
		isoTimeStr := nowBkk.Format(time.RFC3339)

		result := make([]domain.WeatherSensor, 0, len(pattayaList))
		for _, meta := range pattayaList {
			if live, found := liveMap[meta.Serial]; found {
				// Enrich with precise metadata
				live.Order = meta.Order
				live.PoleCode = meta.PoleCode
				live.AdminZone = meta.AdminZone
				live.Subdistrict = meta.Subdistrict
				live.District = meta.District
				if meta.ActualLat != 0 && meta.ActualLng != 0 {
					live.Latitude = meta.ActualLat
					live.Longitude = meta.ActualLng
				}
				if filter.AdminZone != "" && live.AdminZone != filter.AdminZone {
					continue
				}
				result = append(result, live)
			} else {
				// Offline Fallback Entry
				lat := meta.ActualLat
				lng := meta.ActualLng
				if lat == 0 || lng == 0 {
					lat = meta.PlannedLat
					lng = meta.PlannedLng
				}

				if filter.AdminZone != "" && meta.AdminZone != filter.AdminZone {
					continue
				}

				result = append(result, domain.WeatherSensor{
					ID:                meta.Serial,
					Serial:            meta.Serial,
					StationID:         meta.Serial,
					HardwareIMEI:      meta.Serial,
					PoleCode:          meta.PoleCode,
					Order:             meta.Order,
					Name:              fmt.Sprintf("เสาตรวจวัด %s (%s)", meta.PoleCode, meta.AdminZone),
					LocationName:      fmt.Sprintf("%s ต.%s อ.%s", meta.AdminZone, meta.Subdistrict, meta.District),
					SiteName:          fmt.Sprintf("%s ต.%s อ.%s", meta.AdminZone, meta.Subdistrict, meta.District),
					Subdistrict:       meta.Subdistrict,
					District:          meta.District,
					AdminZone:         meta.AdminZone,
					Latitude:          lat,
					Longitude:         lng,
					Temp:              0,
					Temperature:       0,
					Humidity:          0,
					RainRate:          0,
					RainDaily:         0,
					WindSpeed:         0,
					WindDirection:     0,
					WindDirectionText: "ไม่มีข้อมูล",
					PM25:              0,
					Pressure:          1010,
					AirPressure:       1010,
					Status:            "normal",
					OnlineStatus:      "offline",
					TimeOnsite:        timeStr,
					LastUpdated:       timeStr,
					UpdatedAt:         isoTimeStr,
				})
			}
		}

		return result, nil
	}

	// Scope: All (Default for Web Admin)
	result := make([]domain.WeatherSensor, 0, len(allSensors))
	searchQuery := strings.ToLower(strings.TrimSpace(filter.Search))

	for _, s := range allSensors {
		if filter.District != "" && s.District != filter.District {
			continue
		}
		if filter.AdminZone != "" && s.AdminZone != filter.AdminZone {
			continue
		}
		if filter.Status != "" && filter.Status != "all" && s.Status != filter.Status {
			continue
		}
		if searchQuery != "" {
			combined := strings.ToLower(fmt.Sprintf("%s %s %s %s %s %s",
				s.Serial, s.PoleCode, s.Name, s.LocationName, s.Subdistrict, s.District))
			if !strings.Contains(combined, searchQuery) {
				continue
			}
		}
		result = append(result, s)
	}

	return result, nil
}

func (u *weatherUseCase) GetSensorBySerial(ctx context.Context, serial string) (*domain.WeatherSensor, error) {
	sensors, err := u.repo.FetchAllSensors(ctx)
	if err != nil {
		return nil, err
	}

	for _, s := range sensors {
		if s.Serial == serial || s.ID == serial {
			return &s, nil
		}
	}

	// Check if in Pattaya 165 offline list
	if meta, found := u.repo.GetPattaya165Meta(serial); found {
		lat := meta.ActualLat
		lng := meta.ActualLng
		if lat == 0 || lng == 0 {
			lat = meta.PlannedLat
			lng = meta.PlannedLng
		}
		nowBkk := time.Now().In(time.FixedZone("Asia/Bangkok", 7*3600))
		return &domain.WeatherSensor{
			ID:                meta.Serial,
			Serial:            meta.Serial,
			StationID:         meta.Serial,
			HardwareIMEI:      meta.Serial,
			PoleCode:          meta.PoleCode,
			Order:             meta.Order,
			Name:              fmt.Sprintf("เสาตรวจวัด %s (%s)", meta.PoleCode, meta.AdminZone),
			LocationName:      fmt.Sprintf("%s ต.%s อ.%s", meta.AdminZone, meta.Subdistrict, meta.District),
			SiteName:          fmt.Sprintf("%s ต.%s อ.%s", meta.AdminZone, meta.Subdistrict, meta.District),
			Subdistrict:       meta.Subdistrict,
			District:          meta.District,
			AdminZone:         meta.AdminZone,
			Latitude:          lat,
			Longitude:         lng,
			Status:            "normal",
			OnlineStatus:      "offline",
			LastUpdated:       nowBkk.Format("15:04 น."),
			UpdatedAt:         nowBkk.Format(time.RFC3339),
		}, nil
	}

	return nil, fmt.Errorf("station not found: %s", serial)
}

func (u *weatherUseCase) GetDashboardStats(ctx context.Context) (*domain.WeatherDashboardStats, error) {
	sensors, err := u.repo.FetchAllSensors(ctx)
	if err != nil {
		return nil, err
	}

	total := len(sensors)
	online := 0
	silent := 0
	malfunctions := 0

	for _, s := range sensors {
		if s.OnlineStatus == "online" {
			online++
		} else {
			silent++
		}
		if s.Status == "danger" {
			malfunctions++
		}
	}

	nowBkk := time.Now().In(time.FixedZone("Asia/Bangkok", 7*3600))
	return &domain.WeatherDashboardStats{
		TotalRegisters: total,
		OnlineDevices:  online,
		SilentDevices:  silent,
		Malfunctions:   malfunctions,
		SyncTime:       nowBkk.Format("2006-01-02 15:04:05"),
	}, nil
}

func (u *weatherUseCase) GetMapFrames(ctx context.Context, variable, animType, date string) ([]domain.WeatherTileFrame, error) {
	if animType == "" {
		animType = "nrt"
	}

	upstreamVar := "nrtrr"
	if animType == "nrt" {
		switch variable {
		case "cloud", "nrttb":
			upstreamVar = "nrttb"
		case "rain_daily", "nrtrr_daily":
			upstreamVar = "nrtrr_daily"
		case "rain_hourly", "nrtrr_hourly":
			upstreamVar = "nrtrr_hourly"
		case "wind", "nrtwind":
			upstreamVar = "nrtwind"
		case "temp", "nrttemp":
			upstreamVar = "nrttemp"
		default:
			if strings.HasPrefix(variable, "nrt") {
				upstreamVar = variable
			} else {
				upstreamVar = "nrtrr"
			}
		}
	} else {
		// forecast
		switch variable {
		case "temp":
			upstreamVar = "temp"
		case "wind":
			upstreamVar = "wind"
		case "humidity", "rh":
			upstreamVar = "rh"
		default:
			upstreamVar = "rr"
		}
	}

	limit := 15
	if animType == "forecast" || variable == "rain_hourly" {
		limit = 24
	}

	if (animType == "forecast" || date == "") && date == "" {
		nowBkk := time.Now().In(time.FixedZone("Asia/Bangkok", 7*3600))
		date = nowBkk.Format("2006-01-02")
	}

	return u.repo.FetchMapFrames(ctx, upstreamVar, animType, date, limit)
}
