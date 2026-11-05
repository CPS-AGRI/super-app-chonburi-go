package repository

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"super-app-chonburi-go/config"
	"super-app-chonburi-go/internal/domain"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

//go:embed assets/pattaya_165_stations.json
var pattaya165RawJSON []byte

const (
	redisKeyWeatherSensorsAll = "chonburi:weather:sensors:all"
	redisTTLWeatherSensorsAll = 60 * time.Second
	l1TTLWeatherSensorsAll    = 15 * time.Second

	redisPrefixWeatherFrames = "chonburi:weather:frames:"
	redisTTLWeatherFramesNRT = 60 * time.Second
	redisTTLWeatherFramesFC  = 10 * time.Minute
)

var windCompassDirections = [...]string{
	"เหนือ (N)",
	"ต.อ.น. (NE)",
	"ตะวันออก (E)",
	"ต.อ.ต. (SE)",
	"ใต้ (S)",
	"ต.ต.ต. (SW)",
	"ตะวันตก (W)",
	"ต.ต.น. (NW)",
}

var thaiMonthsShortMap = [...]string{
	"", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.",
	"ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค.",
}

type WeatherRepository interface {
	GetActiveAccounts(ctx context.Context) ([]domain.WeatherAccount, error)
	FetchAllSensors(ctx context.Context) ([]domain.WeatherSensor, error)
	FetchMapFrames(ctx context.Context, variable, animType, date string, limit int) ([]domain.WeatherTileFrame, error)
	GetPattaya165List() []domain.PattayaStationMeta
	GetPattaya165Meta(serial string) (domain.PattayaStationMeta, bool)
	FlushCache(ctx context.Context)

	// Alert Settings (Real Database in PostgreSQL)
	GetAlertSettings(ctx context.Context) ([]domain.WeatherAlertSetting, error)
	CreateAlertSetting(ctx context.Context, item *domain.WeatherAlertSetting) error
	UpdateAlertSetting(ctx context.Context, id string, updates map[string]interface{}) (*domain.WeatherAlertSetting, error)
	DeleteAlertSetting(ctx context.Context, id string) error

	// Card Templates
	GetCardTemplates(ctx context.Context) (map[string]string, error)
	UpdateCardTemplate(ctx context.Context, severity, backgroundURL, updatedBy string) error

	// Station Metadata & Custom Aliases
	GetStationMetadata(ctx context.Context, serial string) (*domain.WeatherStationMetadata, error)
	GetAllStationMetadata(ctx context.Context) (map[string]domain.WeatherStationMetadata, error)
	UpsertStationMetadata(ctx context.Context, meta *domain.WeatherStationMetadata) error

	// Historical Telemetry Data from TimescaleDB
	FetchStationHistory(ctx context.Context, serial, imei, rangeType, date, startDate, endDate string) ([]domain.TelemetryHistoryRecord, error)
}

type weatherRepository struct {
	db         *gorm.DB
	cfg        *config.Config
	rdb        *redis.Client
	httpClient *http.Client
	sf         singleflight.Group

	pattaya165List []domain.PattayaStationMeta
	pattaya165Map  map[string]domain.PattayaStationMeta

	// L1 RAM Cache for Weather Sensors
	l1Mu       sync.RWMutex
	l1Cache    []domain.WeatherSensor
	l1CachedAt time.Time

	// Spatial Resolver: tracks serials already resolved to subdistrict (resolve once per serial lifetime)
	resolvedMu    sync.RWMutex
	resolvedCache map[string]bool // serial -> already resolved & upserted
}

func NewWeatherRepository(db *gorm.DB, cfg *config.Config, rdb *redis.Client) WeatherRepository {
	tr := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression: false,
	}

	repo := &weatherRepository{
		db:  db,
		cfg: cfg,
		rdb: rdb,
		httpClient: &http.Client{
			Timeout:   8 * time.Second,
			Transport: tr,
		},
		pattaya165Map: make(map[string]domain.PattayaStationMeta),
		resolvedCache: make(map[string]bool),
	}

	// Parse embedded Pattaya 165 Registry into memory (Zero disk I/O, O(1) lookup)
	if len(pattaya165RawJSON) > 0 {
		var list []domain.PattayaStationMeta
		if err := json.Unmarshal(pattaya165RawJSON, &list); err == nil {
			repo.pattaya165List = list
			for _, item := range list {
				repo.pattaya165Map[item.Serial] = item
			}
			log.Printf("✅ [Weather Repo] Embedded Pattaya 165 Registry loaded successfully (%d stations)", len(list))
		} else {
			log.Printf("⚠️ [Weather Repo] Failed to parse embedded Pattaya 165 JSON: %v", err)
		}
	}

	return repo
}

func (r *weatherRepository) GetPattaya165List() []domain.PattayaStationMeta {
	return r.pattaya165List
}

func (r *weatherRepository) GetPattaya165Meta(serial string) (domain.PattayaStationMeta, bool) {
	meta, exists := r.pattaya165Map[serial]
	return meta, exists
}

func (r *weatherRepository) FlushCache(ctx context.Context) {
	r.l1Mu.Lock()
	r.l1Cache = nil
	r.l1CachedAt = time.Time{}
	r.l1Mu.Unlock()

	if r.rdb != nil {
		_ = r.rdb.Del(ctx, redisKeyWeatherSensorsAll).Err()
	}
}

func (r *weatherRepository) GetActiveAccounts(ctx context.Context) ([]domain.WeatherAccount, error) {
	if r.db == nil {
		return nil, nil
	}
	var accounts []domain.WeatherAccount
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at ASC").
		Find(&accounts).Error
	return accounts, err
}

func getWindCompassText(deg float64) string {
	idx := int(math.Round(math.Mod(deg, 360)/45)) % 8
	if idx < 0 {
		idx += 8
	}
	return windCompassDirections[idx]
}

// FetchAllSensors pulls live data from Fahfon Core Service (10.1.2.25:50059)
// using Bounded Concurrent Goroutines across registered Weather Accounts.
func (r *weatherRepository) FetchAllSensors(ctx context.Context) ([]domain.WeatherSensor, error) {
	// 1. Check L1 Memory Cache (< 0.2ms)
	r.l1Mu.RLock()
	if r.l1Cache != nil && time.Since(r.l1CachedAt) < l1TTLWeatherSensorsAll {
		cached := r.l1Cache
		r.l1Mu.RUnlock()
		return cached, nil
	}
	r.l1Mu.RUnlock()

	// 2. Check L2 Redis Cache (< 2ms)
	if r.rdb != nil {
		val, err := r.rdb.Get(ctx, redisKeyWeatherSensorsAll).Result()
		if err == nil && val != "" {
			var list []domain.WeatherSensor
			if err := json.Unmarshal([]byte(val), &list); err == nil && len(list) > 0 {
				r.l1Mu.Lock()
				r.l1Cache = list
				r.l1CachedAt = time.Now()
				r.l1Mu.Unlock()
				return list, nil
			}
		}
	}

	// 3. SingleFlight: Prevent Thundering Herd / Cache Stampede
	val, err, _ := r.sf.Do("fetch_chonburi_core_stations", func() (any, error) {
		accounts, err := r.GetActiveAccounts(ctx)
		if err != nil {
			log.Printf("⚠️ [Weather Repo] Failed to load active accounts from DB: %v", err)
		}

		type accountResult struct {
			uid   string
			name  string
			items []domain.FahfonStationItem
			err   error
		}

		var rawItems []domain.FahfonStationItem
		nowBkk := time.Now().In(time.FixedZone("Asia/Bangkok", 7*3600))
		timeStr := nowBkk.Format("15:04 น.")
		isoTimeStr := nowBkk.Format(time.RFC3339)

		if len(accounts) > 0 {
			// Concurrently query each account via Bounded Semaphore (8 Workers)
			resultsChan := make(chan accountResult, len(accounts))
			var wg sync.WaitGroup
			sem := make(chan struct{}, 8)

			for _, acc := range accounts {
				wg.Add(1)
				go func(account domain.WeatherAccount) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()

					items, fetchErr := r.fetchStationsForAccount(ctx, account.UniqueID)
					resultsChan <- accountResult{
						uid:   account.UniqueID,
						name:  account.AccountName,
						items: items,
						err:   fetchErr,
					}
				}(acc)
			}

			wg.Wait()
			close(resultsChan)

			for res := range resultsChan {
				if res.err != nil {
					log.Printf("⚠️ [Weather Repo] Account %s (%s) fetch error: %v", res.uid, res.name, res.err)
					continue
				}
				rawItems = append(rawItems, res.items...)
			}
		}

		// Fallback: If no accounts configured or accounts returned 0 items, query global stations endpoint
		if len(rawItems) == 0 {
			globalItems, fallbackErr := r.fetchGlobalStations(ctx)
			if fallbackErr == nil && len(globalItems) > 0 {
				rawItems = globalItems
			}
		}

		// Load custom station metadata/aliases from Chonburi DB
		customMetaMap := make(map[string]domain.WeatherStationMetadata)
		if r.db != nil {
			var metas []domain.WeatherStationMetadata
			if err := r.db.WithContext(ctx).Find(&metas).Error; err == nil {
				for _, m := range metas {
					customMetaMap[m.Serial] = m
				}
			}
		}

		// Normalize & Deduplicate by Serial/IMEI
		sensorMap := make(map[string]domain.WeatherSensor)
		for _, it := range rawItems {
			serial := strings.TrimSpace(it.Serial)
			if serial == "" {
				serial = strings.TrimSpace(it.RegistedIMEI.Serial)
			}
			if serial == "" {
				serial = strings.TrimSpace(it.IMEI)
			}
			if serial == "" {
				continue
			}

			lat, _ := strconv.ParseFloat(it.Position.Latitude, 64)
			lng, _ := strconv.ParseFloat(it.Position.Longitude, 64)
			if lat == 0 || lng == 0 {
				lat, _ = strconv.ParseFloat(it.RegistedIMEI.Latitude, 64)
				lng, _ = strconv.ParseFloat(it.RegistedIMEI.Longitude, 64)
			}
			if lat == 0 || lng == 0 {
				continue
			}

			// Parameters
			deg := it.Parameters.WindDirection
			rainRate := it.Parameters.Rainfall
			if rainRate == 0 {
				rainRate = it.Parameters.HourlyAccumRain
			}
			rainDaily := it.Parameters.DailyAccumRain
			pm25 := it.Parameters.PM2dot5
			temp := it.Parameters.Temp
			humidity := it.Parameters.RelativeHumidity
			windSpeed := it.Parameters.WindSpeed
			pressure := it.Parameters.BarometricPressure
			if pressure == 0 {
				pressure = 1010
			}

			// Status calculation based on Thailand AQI standard
			status := "normal"
			if pm25 > 75.0 || rainRate > 50.0 {
				status = "danger"
			} else if pm25 > 37.5 || rainRate > 25.0 {
				status = "warning"
			}

			// Station Name & Location resolution
			deviceName := it.StationName
			if deviceName == "" {
				deviceName = it.RegistedIMEI.StationName
			}
			if deviceName == "" {
				deviceName = it.RegistedIMEI.CommonSiteName
			}
			if deviceName == "" {
				deviceName = fmt.Sprintf("สถานีตรวจวัดสภาพอากาศ %s", serial)
			}

			siteName := it.RegistedIMEI.CommonSiteName
			if siteName == "" {
				siteName = deviceName
			}

			// --- Spatial Resolution: Lat/Lon → Subdistrict/District ---
			// Priority: 1) customMeta (Admin override) 2) Pattaya 165 file 3) Spatial Resolver (Auto)

			poleCode := ""
			adminZone := ""
			subdistrict := "ชลบุรี"
			district := "ชลบุรี"
			order := 0

			// Priority 2: Pattaya 165 Registry (backward compat — only for matched serials)
			if meta, hasMeta := r.GetPattaya165Meta(serial); hasMeta {
				poleCode = meta.PoleCode
				adminZone = meta.AdminZone
				subdistrict = meta.Subdistrict
				district = meta.District
				order = meta.Order
				deviceName = fmt.Sprintf("เสาตรวจวัด %s (%s)", poleCode, adminZone)
				siteName = fmt.Sprintf("%s ต.%s อ.%s", adminZone, subdistrict, district)
			}

			// Priority 1: Custom Alias / Landmark Override from Chonburi DB (highest priority)
			if customMeta, hasCustom := customMetaMap[serial]; hasCustom && customMeta.CustomName != "" {
				deviceName = customMeta.CustomName
				if customMeta.PoleCode != "" {
					poleCode = customMeta.PoleCode
				}
				if customMeta.Subdistrict != "" {
					subdistrict = customMeta.Subdistrict
				}
				if customMeta.District != "" {
					district = customMeta.District
				}
				if customMeta.AdminZone != "" {
					adminZone = customMeta.AdminZone
				}
				if customMeta.LandmarkDesc != "" {
					siteName = fmt.Sprintf("%s (%s ต.%s อ.%s)", customMeta.LandmarkDesc, adminZone, subdistrict, district)
				}
			} else if subdistrict == "ชลบุรี" && lat != 0 && lng != 0 {
				// Priority 3: Spatial Auto-Resolve — only when subdistrict still unknown
				// Check if this serial has already been resolved to avoid repeated UPSERT
				r.resolvedMu.RLock()
				alreadyResolved := r.resolvedCache[serial]
				r.resolvedMu.RUnlock()

				resolvedSub, resolvedDist := globalSpatialResolver.Resolve(lat, lng)
				subdistrict = resolvedSub
				district = resolvedDist

				if !alreadyResolved && r.db != nil {
					// Fire-and-forget UPSERT into weather_station_metadata (non-blocking)
					go func(s, sub, dist string) {
						upsertMeta := &domain.WeatherStationMetadata{
							Serial:      s,
							Subdistrict: sub,
							District:    dist,
						}
						result := r.db.Where("serial = ?", s).FirstOrCreate(upsertMeta)
						if result.Error == nil {
							// If record existed but subdistrict was blank, patch it
							if upsertMeta.Subdistrict == "" || upsertMeta.Subdistrict == "ชลบุรี" {
								r.db.Model(upsertMeta).Where("serial = ?", s).Updates(map[string]interface{}{
									"subdistrict": sub,
									"district":    dist,
								})
							}
							log.Printf("✅ [SpatialResolver] %s → ต.%s อ.%s (auto-mapped from Lat/Lon)", s, sub, dist)
						} else {
							log.Printf("⚠️ [SpatialResolver] UPSERT failed for %s: %v", s, result.Error)
						}
					}(serial, resolvedSub, resolvedDist)

					r.resolvedMu.Lock()
					r.resolvedCache[serial] = true
					r.resolvedMu.Unlock()
				}
			}

			updatedTime := timeStr
			if it.Position.TimeStampUTC > 0 {
				updatedTime = time.Unix(it.Position.TimeStampUTC, 0).In(time.FixedZone("Asia/Bangkok", 7*3600)).Format("15:04 น.")
			}

			sensorMap[serial] = domain.WeatherSensor{
				ID:                serial,
				Serial:            serial,
				StationID:         serial,
				HardwareIMEI:      it.IMEI,
				PoleCode:          poleCode,
				Order:             order,
				Name:              deviceName,
				LocationName:      siteName,
				SiteName:          siteName,
				Subdistrict:       subdistrict,
				District:          district,
				AdminZone:         adminZone,
				Latitude:          lat,
				Longitude:         lng,
				Temp:              temp,
				Temperature:       temp,
				Humidity:          humidity,
				RainRate:          rainRate,
				RainHourly:        it.Parameters.HourlyAccumRain,
				RainDaily:         rainDaily,
				WindSpeed:         windSpeed,
				WindDirection:     deg,
				WindDirectionText: getWindCompassText(deg),
				PM1:               it.Parameters.PM1dot0,
				PM25:              pm25,
				PM4:               it.Parameters.PM4,
				PM10:              it.Parameters.PM10,
				CO2:               it.Parameters.CO2,
				UVIndex:           it.Parameters.UVIndex,
				LightIntensity:    it.Parameters.LightIntensity,
				Pressure:          pressure,
				AirPressure:       pressure,
				Status:            status,
				OnlineStatus:      "online",
				TimeOnsite:        updatedTime,
				LastUpdated:       updatedTime,
				UpdatedAt:         isoTimeStr,
			}
		}

		sensors := make([]domain.WeatherSensor, 0, len(sensorMap))
		for _, s := range sensorMap {
			sensors = append(sensors, s)
		}

		// Update L1 Cache
		r.l1Mu.Lock()
		r.l1Cache = sensors
		r.l1CachedAt = time.Now()
		r.l1Mu.Unlock()

		// Update L2 Redis Cache
		if r.rdb != nil && len(sensors) > 0 {
			go func(data []domain.WeatherSensor) {
				if b, err := json.Marshal(data); err == nil {
					_ = r.rdb.Set(context.Background(), redisKeyWeatherSensorsAll, b, redisTTLWeatherSensorsAll).Err()
				}
			}(sensors)
		}

		return sensors, nil
	})

	if err != nil {
		return nil, err
	}
	return val.([]domain.WeatherSensor), nil
}

func (r *weatherRepository) fetchStationsForAccount(ctx context.Context, accountID string) ([]domain.FahfonStationItem, error) {
	baseURL := strings.TrimRight(r.cfg.FahfonWeatherBaseURL, "/")
	reqURL := fmt.Sprintf("%s/api/v1/weather/stations?account_id=%s", baseURL, url.QueryEscape(accountID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-service-key", r.cfg.FahfonServiceKey)
	req.Header.Set("Accept", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upstream core error status %d: %s", resp.StatusCode, string(body))
	}

	var raw domain.FahfonStationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return raw.Items, nil
}

func (r *weatherRepository) fetchGlobalStations(ctx context.Context) ([]domain.FahfonStationItem, error) {
	baseURL := strings.TrimRight(r.cfg.FahfonWeatherBaseURL, "/")
	reqURL := fmt.Sprintf("%s/api/v1/weather/stations", baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-service-key", r.cfg.FahfonServiceKey)
	req.Header.Set("Accept", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("global stations status %d", resp.StatusCode)
	}

	var raw domain.FahfonStationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err == nil && len(raw.Items) > 0 {
		return raw.Items, nil
	}
	return nil, fmt.Errorf("no stations returned")
}

// FetchMapFrames proxies and caches radar/satellite frames directly from Core Service (10.1.2.25:50059)
func (r *weatherRepository) FetchMapFrames(ctx context.Context, variable, animType, date string, limit int) ([]domain.WeatherTileFrame, error) {
	cacheKey := fmt.Sprintf("%s%s:%s:%s:%d", redisPrefixWeatherFrames, variable, animType, date, limit)

	// Check Redis cache
	if r.rdb != nil {
		if cached, err := r.rdb.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
			var frames []domain.WeatherTileFrame
			if err := json.Unmarshal([]byte(cached), &frames); err == nil && len(frames) > 0 {
				return frames, nil
			}
		}
	}

	val, err, _ := r.sf.Do(cacheKey, func() (any, error) {
		baseURL := strings.TrimRight(r.cfg.FahfonWeatherBaseURL, "/")
		reqURL := fmt.Sprintf("%s/api/v1/weather/map/frames?variable=%s&anim_type=%s&limit=%d",
			baseURL, variable, animType, limit)
		if date != "" {
			reqURL += "&date=" + date
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("x-service-key", r.cfg.FahfonServiceKey)
		req.Header.Set("Accept", "application/json")

		resp, err := r.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("upstream core map frames call failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("upstream core map frames status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
		if err != nil {
			return nil, err
		}

		var jsonResp struct {
			Data struct {
				Frames []map[string]any `json:"frames"`
			} `json:"data"`
			Frames []map[string]any `json:"frames"`
		}
		_ = json.Unmarshal(body, &jsonResp)

		rawFrames := jsonResp.Data.Frames
		if len(rawFrames) == 0 {
			rawFrames = jsonResp.Frames
		}

		frames := make([]domain.WeatherTileFrame, 0, len(rawFrames))
		for _, f := range rawFrames {
			formatted, _ := f["formatted"].(string)
			compositeURL, _ := f["composite_url"].(string)
			if compositeURL == "" {
				compositeURL, _ = f["url"].(string)
			}
			validTime, _ := f["valid_time"].(string)

			displayTime := formatFrameTime(formatted)

			frames = append(frames, domain.WeatherTileFrame{
				TimeLabel:   displayTime,
				DisplayTime: displayTime,
				Formatted:   formatted,
				URL:         compositeURL,
				ValidTime:   validTime,
			})
		}

		// Cache in Redis
		if r.rdb != nil && len(frames) > 0 {
			ttl := redisTTLWeatherFramesNRT
			if animType == "forecast" {
				ttl = redisTTLWeatherFramesFC
			}
			if b, err := json.Marshal(frames); err == nil {
				_ = r.rdb.Set(context.Background(), cacheKey, b, ttl).Err()
			}
		}

		return frames, nil
	})

	if err != nil {
		return nil, err
	}
	return val.([]domain.WeatherTileFrame), nil
}

func formatFrameTime(formatted string) string {
	if len(formatted) == 8 {
		dNum, _ := strconv.Atoi(formatted[6:8])
		mNum, _ := strconv.Atoi(formatted[4:6])
		mLabel := ""
		if mNum >= 1 && mNum <= 12 {
			mLabel = thaiMonthsShortMap[mNum]
		}
		return fmt.Sprintf("%d %s", dNum, mLabel)
	} else if len(formatted) >= 12 {
		hh := formatted[8:10]
		mm := formatted[10:12]
		return fmt.Sprintf("%s:%s น.", hh, mm)
	}
	return formatted
}

// --- Alert Settings CRUD (PostgreSQL) ---

func (r *weatherRepository) GetAlertSettings(ctx context.Context) ([]domain.WeatherAlertSetting, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection unavailable")
	}
	var settings []domain.WeatherAlertSetting
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&settings).Error
	return settings, err
}

func (r *weatherRepository) CreateAlertSetting(ctx context.Context, item *domain.WeatherAlertSetting) error {
	if r.db == nil {
		return fmt.Errorf("database connection unavailable")
	}
	if item.ID == "" {
		item.ID = fmt.Sprintf("W-%03d", time.Now().UnixNano()%1000)
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *weatherRepository) UpdateAlertSetting(ctx context.Context, id string, updates map[string]interface{}) (*domain.WeatherAlertSetting, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection unavailable")
	}
	updates["updated_at"] = time.Now()
	var setting domain.WeatherAlertSetting
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&setting).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&setting).Updates(updates).Error; err != nil {
		return nil, err
	}
	r.db.WithContext(ctx).Where("id = ?", id).First(&setting)
	return &setting, nil
}

func (r *weatherRepository) DeleteAlertSetting(ctx context.Context, id string) error {
	if r.db == nil {
		return fmt.Errorf("database connection unavailable")
	}
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.WeatherAlertSetting{}).Error
}

// --- Card Templates Management ---

func (r *weatherRepository) GetCardTemplates(ctx context.Context) (map[string]string, error) {
	minioBase := os.Getenv("MINIO_PUBLIC_BASE_URL")
	if minioBase == "" {
		minioBase = "https://uat-super-app-chonburi-api.mueangsmart.com"
	}
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "super-app-assets"
	}
	baseURL := fmt.Sprintf("%s/%s/weather-templates", strings.TrimRight(minioBase, "/"), bucket)

	result := map[string]string{
		"info":    fmt.Sprintf("%s/bg_info_cinematic.jpg", baseURL),
		"warning": fmt.Sprintf("%s/bg_warning_cinematic.jpg", baseURL),
		"danger":  fmt.Sprintf("%s/bg_danger_cinematic.jpg", baseURL),
	}
	if r.db == nil {
		return result, nil
	}
	var templates []domain.WeatherCardTemplate
	if err := r.db.WithContext(ctx).Find(&templates).Error; err == nil {
		for _, t := range templates {
			if t.BackgroundURL != "" {
				result[t.Severity] = t.BackgroundURL
			}
		}
	}
	return result, nil
}

func (r *weatherRepository) UpdateCardTemplate(ctx context.Context, severity, backgroundURL, updatedBy string) error {
	if r.db == nil {
		return fmt.Errorf("database connection unavailable")
	}
	template := domain.WeatherCardTemplate{
		Severity:      severity,
		BackgroundURL: backgroundURL,
		UpdatedBy:     updatedBy,
		UpdatedAt:     time.Now(),
	}
	return r.db.WithContext(ctx).Save(&template).Error
}

// --- Station Metadata & Custom Aliases ---

func (r *weatherRepository) GetStationMetadata(ctx context.Context, serial string) (*domain.WeatherStationMetadata, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection unavailable")
	}
	var meta domain.WeatherStationMetadata
	err := r.db.WithContext(ctx).Where("serial = ?", serial).First(&meta).Error
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func (r *weatherRepository) GetAllStationMetadata(ctx context.Context) (map[string]domain.WeatherStationMetadata, error) {
	result := make(map[string]domain.WeatherStationMetadata)
	if r.db == nil {
		return result, nil
	}
	var metas []domain.WeatherStationMetadata
	err := r.db.WithContext(ctx).Find(&metas).Error
	if err != nil {
		return result, err
	}
	for _, m := range metas {
		result[m.Serial] = m
	}
	return result, nil
}

func (r *weatherRepository) UpsertStationMetadata(ctx context.Context, meta *domain.WeatherStationMetadata) error {
	if r.db == nil {
		return fmt.Errorf("database connection unavailable")
	}
	meta.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Save(meta).Error
	if err == nil {
		// Invalidate cache immediately so new custom alias is active across all endpoints
		r.FlushCache(ctx)
	}
	return err
}

func (r *weatherRepository) FetchStationHistory(ctx context.Context, serial, imei, rangeType, date, startDate, endDate string) ([]domain.TelemetryHistoryRecord, error) {
	if rangeType == "" {
		rangeType = "hourly"
	}

	targetImei := imei
	if targetImei == "" {
		targetImei = serial
	}

	isRangeOrCustomDate := (startDate != "" && endDate != "") || date != ""

	// Helper: Query FahfonWeatherBaseURL (ff-weather-service at 10.1.2.25:50059)
	fetchFromFahfonWeather := func() ([]domain.TelemetryHistoryRecord, error) {
		if r.cfg.FahfonWeatherBaseURL == "" || targetImei == "" {
			return nil, fmt.Errorf("fahfon weather baseURL or imei empty")
		}
		baseURL := strings.TrimRight(r.cfg.FahfonWeatherBaseURL, "/")
		reqURL := fmt.Sprintf("%s/api/v1/weather/stations/imei/%s/graph", baseURL, url.PathEscape(targetImei))
		params := url.Values{}
		if startDate != "" && endDate != "" {
			params.Set("start_date", startDate)
			params.Set("end_date", endDate)
		} else if date != "" {
			params.Set("date", date)
		}
		if len(params) > 0 {
			reqURL += "?" + params.Encode()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("x-service-key", r.cfg.FahfonServiceKey)
		req.Header.Set("Accept", "application/json")
		resp, err := r.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fahfon weather HTTP status %d", resp.StatusCode)
		}

		var graphResp struct {
			Parameters []struct {
				Time             string  `json:"time"`
				Temp             float64 `json:"temp"`
				RelativeHumidity float64 `json:"relativeHumidity"`
				BarometricPres   float64 `json:"barometricPres"`
				WindSpeed        float64 `json:"windSpeed"`
				WindDirection    float64 `json:"windDirection"`
				RainIntensity    float64 `json:"rainIntensity"`
				RainHourly       float64 `json:"rainHourly"`
				RainDaily        float64 `json:"rainDaily"`
				UV               float64 `json:"uv"`
				PM1              float64 `json:"pm1"`
				PM25             float64 `json:"pm25"`
				PM4              float64 `json:"pm4"`
				PM10             float64 `json:"pm10"`
				CO2              float64 `json:"co2"`
			} `json:"parameters"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&graphResp); err != nil {
			return nil, err
		}
		if len(graphResp.Parameters) == 0 {
			return nil, fmt.Errorf("no parameters in graph response")
		}

		records := make([]domain.TelemetryHistoryRecord, 0, len(graphResp.Parameters))
		for _, p := range graphResp.Parameters {
			records = append(records, domain.TelemetryHistoryRecord{
				Time:          p.Time,
				TimeOnsite:    p.Time,
				Temp:          p.Temp,
				Humid:         p.RelativeHumidity,
				Press:         p.BarometricPres,
				RainRate:      p.RainIntensity,
				RainHourly:    p.RainHourly,
				RainDaily:     p.RainDaily,
				PM1:           p.PM1,
				PM25:          p.PM25,
				PM40:          p.PM4,
				PM10:          p.PM10,
				WindSpeed:     p.WindSpeed,
				WindDirection: p.WindDirection,
				CO2:           p.CO2,
				UV:            p.UV,
			})
		}
		return records, nil
	}

	// Helper: Query RiverServiceURL (IoT Core at 10.1.2.19:8081)
	fetchFromRiverService := func() ([]domain.TelemetryHistoryRecord, error) {
		if r.cfg.RiverServiceURL == "" {
			return nil, fmt.Errorf("river service URL empty")
		}
		baseURL := strings.TrimRight(r.cfg.RiverServiceURL, "/")
		reqURL := fmt.Sprintf("%s/api/v1/fahfon-sense/devices/%s/detail?range=%s", baseURL, url.PathEscape(serial), url.QueryEscape(rangeType))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		if r.cfg.FahfonServiceKey != "" {
			req.Header.Set("x-service-key", r.cfg.FahfonServiceKey)
		}
		resp, err := r.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("river service HTTP status %d", resp.StatusCode)
		}

		var detailResp struct {
			History []struct {
				TimeOnsite     string  `json:"time_onsite"`
				Time           string  `json:"time"`
				BucketTime     string  `json:"bucket_time"`
				Temp           float64 `json:"temp"`
				Humid          float64 `json:"humid"`
				Press          float64 `json:"press"`
				WindSpeed      float64 `json:"wind_speed"`
				WindDirection  float64 `json:"wind_direction"`
				RainRate       float64 `json:"rain_rate"`
				RainHourly     float64 `json:"rain_hourly"`
				RainDaily      float64 `json:"rain_daily"`
				UV             float64 `json:"uv"`
				CO2            float64 `json:"co2"`
				PM1            float64 `json:"pm1"`
				PM25           float64 `json:"pm25"`
				PM40           float64 `json:"pm40"`
				PM10           float64 `json:"pm10"`
				Battery        float64 `json:"battery"`
				SignalStrength float64 `json:"signal_strength"`
			} `json:"history"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&detailResp); err != nil {
			return nil, err
		}
		if len(detailResp.History) == 0 {
			return nil, fmt.Errorf("no history in river service")
		}

		records := make([]domain.TelemetryHistoryRecord, 0, len(detailResp.History))
		for _, h := range detailResp.History {
			timeVal := h.TimeOnsite
			if timeVal == "" {
				timeVal = h.Time
			}
			if timeVal == "" {
				timeVal = h.BucketTime
			}
			records = append(records, domain.TelemetryHistoryRecord{
				Time:           timeVal,
				TimeOnsite:     timeVal,
				Temp:           h.Temp,
				Humid:          h.Humid,
				Press:          h.Press,
				RainRate:       h.RainRate,
				RainHourly:     h.RainHourly,
				RainDaily:      h.RainDaily,
				PM1:            h.PM1,
				PM25:           h.PM25,
				PM40:           h.PM40,
				PM10:           h.PM10,
				WindSpeed:      h.WindSpeed,
				WindDirection:  h.WindDirection,
				CO2:            h.CO2,
				UV:             h.UV,
				Battery:        h.Battery,
				SignalStrength: h.SignalStrength,
			})
		}
		return records, nil
	}

	// 1. If date range or custom date is specified, prioritize FahfonWeatherBaseURL (TimescaleDB)
	if isRangeOrCustomDate {
		if recs, err := fetchFromFahfonWeather(); err == nil && len(recs) > 0 {
			return recs, nil
		}
		if recs, err := fetchFromRiverService(); err == nil && len(recs) > 0 {
			return recs, nil
		}
	} else {
		// 2. Otherwise default to RiverService, fallback to FahfonWeather
		if recs, err := fetchFromRiverService(); err == nil && len(recs) > 0 {
			return recs, nil
		}
		if recs, err := fetchFromFahfonWeather(); err == nil && len(recs) > 0 {
			return recs, nil
		}
	}

	// 3. Upstream Telemetry Service fallback
	upstreamURL := fmt.Sprintf("https://microsis.mueangsmart.com/api/v1/fahfon-sense/devices/%s/detail?range=%s", url.PathEscape(serial), url.QueryEscape(rangeType))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL, nil)
	if err == nil {
		req.Header.Set("Accept", "application/json")
		resp, err := r.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var detailResp struct {
				History []struct {
					TimeOnsite     string  `json:"time_onsite"`
					Time           string  `json:"time"`
					BucketTime     string  `json:"bucket_time"`
					Temp           float64 `json:"temp"`
					Humid          float64 `json:"humid"`
					Press          float64 `json:"press"`
					WindSpeed      float64 `json:"wind_speed"`
					WindDirection  float64 `json:"wind_direction"`
					RainRate       float64 `json:"rain_rate"`
					RainHourly     float64 `json:"rain_hourly"`
					RainDaily      float64 `json:"rain_daily"`
					UV             float64 `json:"uv"`
					CO2            float64 `json:"co2"`
					PM1            float64 `json:"pm1"`
					PM25           float64 `json:"pm25"`
					PM40           float64 `json:"pm40"`
					PM10           float64 `json:"pm10"`
					Battery        float64 `json:"battery"`
					SignalStrength float64 `json:"signal_strength"`
				} `json:"history"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&detailResp); err == nil && len(detailResp.History) > 0 {
				records := make([]domain.TelemetryHistoryRecord, 0, len(detailResp.History))
				for _, h := range detailResp.History {
					timeVal := h.TimeOnsite
					if timeVal == "" {
						timeVal = h.Time
					}
					if timeVal == "" {
						timeVal = h.BucketTime
					}
					records = append(records, domain.TelemetryHistoryRecord{
						Time:           timeVal,
						TimeOnsite:     timeVal,
						Temp:           h.Temp,
						Humid:          h.Humid,
						Press:          h.Press,
						RainRate:       h.RainRate,
						RainHourly:     h.RainHourly,
						RainDaily:      h.RainDaily,
						PM1:            h.PM1,
						PM25:           h.PM25,
						PM40:           h.PM40,
						PM10:           h.PM10,
						WindSpeed:      h.WindSpeed,
						WindDirection:  h.WindDirection,
						CO2:            h.CO2,
						UV:             h.UV,
						Battery:        h.Battery,
						SignalStrength: h.SignalStrength,
					})
				}
				return records, nil
			}
		} else if resp != nil {
			resp.Body.Close()
		}
	}

	return []domain.TelemetryHistoryRecord{}, nil
}
