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

			// Metadata lookup from Pattaya 165 Registry
			meta, hasMeta := r.GetPattaya165Meta(serial)
			poleCode := ""
			adminZone := ""
			subdistrict := "ชลบุรี"
			district := "ชลบุรี"
			order := 0

			if hasMeta {
				poleCode = meta.PoleCode
				adminZone = meta.AdminZone
				subdistrict = meta.Subdistrict
				district = meta.District
				order = meta.Order
				deviceName = fmt.Sprintf("เสาตรวจวัด %s (%s)", poleCode, adminZone)
				siteName = fmt.Sprintf("%s ต.%s อ.%s", adminZone, subdistrict, district)
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
				RainDaily:         rainDaily,
				WindSpeed:         windSpeed,
				WindDirection:     deg,
				WindDirectionText: getWindCompassText(deg),
				PM25:              pm25,
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
