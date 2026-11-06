package domain

import "time"

// WaterStation represents a water level monitoring station formatted for client apps (Mobile, Web Admin, Pattaya Dashboard)
type WaterStation struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	LocationName   string              `json:"locationName"`
	Lat            float64             `json:"lat"`
	Lng            float64             `json:"lng"`
	WaterLevel     float64             `json:"waterLevel"`
	WaterLevelMax  float64             `json:"waterLevelMax"`
	RainAmount     float64             `json:"rainAmount"`
	WindSpeed      float64             `json:"windSpeed"`
	WindDirection  float64             `json:"windDirection"`
	Status         string              `json:"status"` // "normal" | "warning" | "critical" | "emergency"
	OnlineStatus   string              `json:"onlineStatus"` // "online" | "offline"
	LastUpdated    string              `json:"lastUpdated"` // Thai format (e.g., "21 ต.ค. 2026, 11:05 น.")
	UpdatedAt      string              `json:"updatedAt"`   // ISO 8601 UTC/ICT format
	SnapshotURL    string              `json:"snapshotUrl,omitempty"`
	CCTV           string              `json:"cctv,omitempty"`
	History        []WaterHistoryPoint `json:"history,omitempty"`
	Serial         string              `json:"serial"`
	CommonSizeName string              `json:"commonSizeName,omitempty"`
	PctToBank      float64                 `json:"pctToBank"`
	SensorMode     string                  `json:"sensorMode"` // "RIVER" | "SURFACE"
	BottomToBankCM float64                 `json:"bottomToBankCM"`
	BankToSensorCM float64                 `json:"bankToSensorCM,omitempty"`
	NetWaterLevelCM float64                `json:"netWaterLevelCM"`
	Subdistrict    string                  `json:"subdistrict,omitempty"`
	District       string                  `json:"district,omitempty"`
	Thresholds     []WaterStationThreshold `json:"thresholds,omitempty"`
}

// WaterStationThreshold represents an alert threshold configured for a station
type WaterStationThreshold struct {
	AlertLevel   string   `json:"alertLevel"` // "WARNING" | "CRITICAL" | "EMERGENCY"
	ThresholdCM  *float64 `json:"thresholdCM,omitempty"`
	ThresholdPct *float64 `json:"thresholdPct,omitempty"`
	IsActive     bool     `json:"isActive"`
}

// UpdateStationThresholdsRequest represents the payload to update alert thresholds
type UpdateStationThresholdsRequest struct {
	WarningPct   *float64 `json:"warning_pct,omitempty"`
	CriticalPct  *float64 `json:"critical_pct,omitempty"`
	EmergencyPct *float64 `json:"emergency_pct,omitempty"`
	WarningCM    *float64 `json:"warning_cm,omitempty"`
	CriticalCM   *float64 `json:"critical_cm,omitempty"`
	EmergencyCM  *float64 `json:"emergency_cm,omitempty"`
	SensorMode   *string  `json:"sensor_mode,omitempty"`
}

// WaterHistoryPoint represents a historical data point for charting and telemetry table
type WaterHistoryPoint struct {
	Time            string  `json:"time"`
	WaterLevel      float64 `json:"waterLevel"` // meters
	RainAmount      float64 `json:"rainAmount"` // mm
	WindSpeed       float64 `json:"windSpeed"`  // km/h
	HourLabel       string  `json:"hourLabel,omitempty"`
	DateLabel       string  `json:"dateLabel,omitempty"`
	ISODate         string  `json:"isoDate,omitempty"`
	Pct             float64 `json:"pct,omitempty"`
	Status          string  `json:"status,omitempty"`
	RawLevelCM      float64 `json:"rawLevelCM,omitempty"`
	NetWaterLevelCM float64 `json:"netWaterLevelCM,omitempty"`
	FilteredLevelCM float64 `json:"filteredLevelCM,omitempty"`
	ConfidenceScore float64 `json:"confidenceScore,omitempty"`
	FilterStatus    string  `json:"filterStatus,omitempty"`
	SnapshotURL     string  `json:"snapshotUrl,omitempty"`
}

// WaterAlert represents an active flood warning alert
type WaterAlert struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Detail      string `json:"detail"`
	Location    string `json:"location"`
	Subdistrict string `json:"subdistrict,omitempty"`
	District    string `json:"district,omitempty"`
	Time        string `json:"time"`
	Severity    string `json:"severity"` // "normal" | "warning" | "critical" | "emergency"
	Unread      bool   `json:"unread,omitempty"`
}

// WaterSnapshotLog represents an hourly camera snapshot log
type WaterSnapshotLog struct {
	Time       string  `json:"time"`
	ImageURL   string  `json:"imageUrl,omitempty"`
	WaterLevel float64 `json:"waterLevel"`
	Pct        float64 `json:"pct,omitempty"`
}

// WaterDeviceDetail represents full telemetry of a station
type WaterDeviceDetail struct {
	ID              string  `json:"id"`
	Serial          string  `json:"serial"`
	Name            string  `json:"name"`
	Location        string  `json:"location"`
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	Status          string  `json:"status"`
	OnlineStatus    string  `json:"onlineStatus"` // "online" | "offline"
	WaterLevel      float64 `json:"waterLevel"`
	WaterLevelMax   float64 `json:"waterLevelMax"`
	RawLevelCM      float64 `json:"rawLevelCM,omitempty"`
	NetWaterLevelCM float64 `json:"netWaterLevelCM,omitempty"`
	BottomToBankCM  float64 `json:"bottomToBankCM,omitempty"`
	BankToSensorCM  float64 `json:"bankToSensorCM,omitempty"`
	LastUpdated     string  `json:"lastUpdated"`
	UpdatedAt       string  `json:"updatedAt"` // ISO 8601
	Battery         float64 `json:"battery,omitempty"`
	Solar           float64 `json:"solar,omitempty"`
	CCTVStreamURL   string  `json:"cctvStreamUrl,omitempty"`
	SnapshotURL     string  `json:"snapshotUrl,omitempty"`
	CCTV            string                  `json:"cctv,omitempty"`
	PctToBank       float64                 `json:"pctToBank,omitempty"`
	SensorMode      string                  `json:"sensorMode,omitempty"` // "RIVER" | "SURFACE"
	Subdistrict     string                  `json:"subdistrict,omitempty"`
	District        string                  `json:"district,omitempty"`
	Thresholds      []WaterStationThreshold `json:"thresholds,omitempty"`
}

// RiverStationMetadata represents persistent subdistrict, district and landmark metadata in Chonburi DB
type RiverStationMetadata struct {
	Serial       string    `gorm:"primaryKey;column:serial" json:"serial"`
	PoleCode     string    `gorm:"column:pole_code" json:"pole_code"`
	CustomName   string    `gorm:"column:custom_name" json:"custom_name"`
	LandmarkDesc string    `gorm:"column:landmark_desc" json:"landmark_desc"`
	Subdistrict  string    `gorm:"column:subdistrict" json:"subdistrict"`
	District     string    `gorm:"column:district" json:"district"`
	AdminZone    string    `gorm:"column:admin_zone" json:"admin_zone"`
	Latitude     float64   `gorm:"column:latitude" json:"latitude"`
	Longitude    float64   `gorm:"column:longitude" json:"longitude"`
	UpdatedBy    string    `gorm:"column:updated_by" json:"updated_by"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (RiverStationMetadata) TableName() string {
	return "river_station_metadata"
}

// WaterCardTemplate represents background wallpapers for 3 alert severities
type WaterCardTemplate struct {
	Severity      string    `gorm:"primaryKey;column:severity" json:"severity"` // info, warning, danger
	BackgroundURL string    `gorm:"column:background_url;not null" json:"background_url"`
	UpdatedBy     string    `gorm:"column:updated_by" json:"updated_by"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (WaterCardTemplate) TableName() string {
	return "water_card_templates"
}

// WaterStreamResult represents RTSP / HLS / WebRTC stream response
type WaterStreamResult struct {
	Success    bool   `json:"success"`
	StreamURL  string `json:"streamUrl,omitempty"`
	WSURL      string `json:"wsUrl,omitempty"`
	SDPOffer   string `json:"sdp_offer,omitempty"`
	SessionID  string `json:"session_id,omitempty"`
	IceServers string `json:"ice_servers,omitempty"`
	Message    string `json:"message,omitempty"`
}

// GatewayRiverDevice represents raw device object from ff-river-services
type GatewayRiverDevice struct {
	Serial           string  `json:"serial"`
	Name             string  `json:"name"`
	Location         string  `json:"location"`
	InstallLatitude  any     `json:"install_latitude"`
	InstallLongitude any     `json:"install_longitude"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	NetWaterLevelCM  float64 `json:"net_water_level_cm"`
	BottomToBankCM   float64 `json:"bottom_to_bank_cm"`
	WaterLevelPct    float64 `json:"water_level_pct"`
	WarningStatus    string  `json:"warning_status"`
	Status           string  `json:"status"`
	LastSeen         string  `json:"last_seen"`
	RainAmount       float64 `json:"rain_amount"`
	WindSpeed        float64 `json:"wind_speed"`
	WindDirection    float64 `json:"wind_direction"`
	CCTVStreamURL    string  `json:"cctv_stream_url"`
	SnapshotURL0     string  `json:"snapshot_url0"`
	SnapshotURL1     string  `json:"snapshot_url1"`
	CCTVPath0        string  `json:"cctv_path_0"`
	CCTVPath1        string  `json:"cctv_path_1"`
	TimestampUTC     int64   `json:"timestamp_utc"`
}

// GatewayHistoryPoint represents stats point from ff-river-services
type GatewayHistoryPoint struct {
	Time            string   `json:"time"`
	LevelCM         float64  `json:"level_cm"`
	NetWaterLevelCM float64  `json:"net_water_level_cm"`
	WaterLevelPct   float64  `json:"water_level_pct"`
	FilteredLevelCM *float64 `json:"filtered_level_cm,omitempty"`
	ConfidenceScore *float64 `json:"confidence_score,omitempty"`
	FilterStatus    string   `json:"filter_status,omitempty"`
	RainAmount      float64  `json:"rain_amount"`
	WindSpeed       float64  `json:"wind_speed"`
	SnapshotURL0    string   `json:"snapshot_url0"`
	SnapshotURL1    string   `json:"snapshot_url1,omitempty"`
	VoltIn          *float64 `json:"volt_in,omitempty"`
}

// GatewaySnapshotLog represents snapshots from ff-river-services
type GatewaySnapshotLog struct {
	Time            string  `json:"time"`
	NetWaterLevelCM float64 `json:"net_water_level_cm"`
	WaterLevelPct   float64 `json:"water_level_pct"`
	WarningStatus   string  `json:"warning_status"`
	ImageURL        string  `json:"image_url"`
}

// GatewayDeviceDetailEnvelope represents response of /api/v1/river/devices/:serial/detail
type GatewayDeviceDetailEnvelope struct {
	Device struct {
		Serial           string  `json:"serial"`
		Name             string  `json:"name"`
		Status           string  `json:"status"`
		InstallLatitude  any     `json:"install_latitude"`
		InstallLongitude any     `json:"install_longitude"`
		WarningStatus    string  `json:"warning_status"`
		OrganizationName string  `json:"organization_name"`
		Description      string  `json:"description"`
		CCTVStreamURL    string  `json:"cctv_stream_url"`
		SnapshotURL0     string  `json:"snapshot_url0"`
		SnapshotURL1     string  `json:"snapshot_url1"`
		CCTVPath0        string  `json:"cctv_path_0"`
		CCTVPath1        string  `json:"cctv_path_1"`
		LevelCM          float64 `json:"level_cm"`
	} `json:"device"`
	Current struct {
		NetWaterLevelCM float64 `json:"net_water_level_cm"`
		WaterLevelPct   float64 `json:"water_level_pct"`
		Time            string  `json:"time"`
		Battery         float64 `json:"battery"`
		Solar           float64 `json:"solar"`
	} `json:"current"`
	Config struct {
		BottomToBankCM float64 `json:"bottom_to_bank_cm"`
		BankToSensorCM float64 `json:"bank_to_sensor_cm"`
		SensorMode     string  `json:"sensor_mode"`
	} `json:"config"`
	Thresholds []struct {
		AlertLevel   string   `json:"alert_level"`
		ThresholdCM  *float64 `json:"threshold_cm"`
		ThresholdPct *float64 `json:"threshold_pct"`
		IsActive     bool     `json:"is_active"`
	} `json:"thresholds"`
}

// GatewayAlertItem represents alert from /api/v1/internal/alerts/active or /api/v1/river/alerts/active
type GatewayAlertItem struct {
	ID         any     `json:"id"`
	Serial     string  `json:"serial"`
	AlertLevel string  `json:"alert_level"`
	AlertType  string  `json:"alert_type"`
	Value      float64 `json:"value"`
	Threshold  float64 `json:"threshold"`
	Message    string  `json:"message"`
	Status     string  `json:"status"`
	Time       string  `json:"time"`
}
