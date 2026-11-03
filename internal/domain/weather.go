package domain

import "time"

// Raw models returned from ff-weather-service (10.1.2.25:50059)

type FahfonStationItem struct {
	IMEI         string             `json:"imei"`
	StationName  string             `json:"stationName"`
	Serial       string             `json:"serial"`
	IsOwner      bool               `json:"isOwner"`
	Parameters   FahfonParameters   `json:"parameters"`
	Position     FahfonPosition     `json:"position"`
	RegistedIMEI FahfonRegistedIMEI `json:"registedIMEI"`
}

type FahfonParameters struct {
	Temp                            float64 `json:"temp"`
	RelativeHumidity                float64 `json:"relativeHumidity"`
	BarometricPressure              float64 `json:"barometricPressure"`
	Rainfall                        float64 `json:"rainfall"`
	HourlyAccumRain                 float64 `json:"hourlyAccumRain"`
	DailyAccumRain                  float64 `json:"dailyAccumRain"`
	WindSpeed                       float64 `json:"windSpeed"`
	WindDirection                   float64 `json:"windDirection"`
	PM1dot0                         float64 `json:"pM1dot0"`
	PM2dot5                         float64 `json:"pM2dot5"`
	PM4                             float64 `json:"pM4"`
	PM10                            float64 `json:"pM10"`
	CO2                             float64 `json:"cO2"`
	UVIndex                         float64 `json:"uvIndex"`
	LightIntensity                  float64 `json:"lightIntensity"`
	RedLightIntensityGain1          float64 `json:"redLightIntensityGain1"`
	GreenLightIntensityGain1        float64 `json:"greenLightIntensityGain1"`
	BlueLightIntensityGain1         float64 `json:"blueLightIntensityGain1"`
	NearInfraredLightIntensityGain1 float64 `json:"nearInfraredLightIntensityGain1"`
}

type FahfonPosition struct {
	Height       string `json:"height"`
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	TimeStampUTC int64  `json:"timeStampUTC"`
}

type FahfonRegistedIMEI struct {
	AccountID      string `json:"accountId"`
	StationName    string `json:"stationName"`
	Serial         string `json:"serial"`
	CommonSiteName string `json:"commonSiteName"`
	Email          string `json:"email"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	Owner          string `json:"owner"`
	RegisType      string `json:"regisType"`
}

type FahfonStationsResponse struct {
	Items      []FahfonStationItem `json:"items"`
	TotalItems int                 `json:"total_items,omitempty"`
}

// PattayaStationMeta represents metadata of a station in the official 165 Pattaya registry
type PattayaStationMeta struct {
	Order       int     `json:"order"`
	PoleCode    string  `json:"pole_code"`
	Serial      string  `json:"serial"`
	Subdistrict string  `json:"subdistrict"`
	District    string  `json:"district"`
	Province    string  `json:"province"`
	AdminZone   string  `json:"admin_zone"`
	PlannedLat  float64 `json:"planned_lat"`
	PlannedLng  float64 `json:"planned_lng"`
	ActualLat   float64 `json:"actual_lat"`
	ActualLng   float64 `json:"actual_lng"`
	DistanceM   float64 `json:"distance_m"`
}

// WeatherSensor represents a standardized weather telemetry device for client applications
// (Pattaya Dashboard, Chonburi Web Admin, Mobile App)
type WeatherSensor struct {
	ID                string  `json:"id"`
	Serial            string  `json:"serial"`
	StationID         string  `json:"stationId,omitempty"`
	HardwareIMEI      string  `json:"hardwareImei,omitempty"`
	PoleCode          string  `json:"poleCode,omitempty"`
	Order             int     `json:"order,omitempty"`
	Name              string  `json:"name"`
	LocationName      string  `json:"locationName"`
	SiteName          string  `json:"siteName,omitempty"`
	Subdistrict       string  `json:"subdistrict"`
	District          string  `json:"district"`
	AdminZone         string  `json:"adminZone,omitempty"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	Temp              float64 `json:"temp"`        // Celsius
	Temperature       float64 `json:"temperature"` // Celsius (Open API alias)
	Humidity          float64 `json:"humidity"`    // %
	RainRate          float64 `json:"rainRate"`          // mm/h
	RainHourly        float64 `json:"rainHourly"`        // mm/h
	RainDaily         float64 `json:"rainDaily"`         // mm
	WindSpeed         float64 `json:"windSpeed"`         // m/s
	WindDirection     float64 `json:"windDirection"`
	WindDirectionText string  `json:"windDirectionText"`
	PM1               float64 `json:"pm1,omitempty"`     // ug/m3
	PM25              float64 `json:"pm25"`              // ug/m3
	PM4               float64 `json:"pm4,omitempty"`     // ug/m3
	PM10              float64 `json:"pm10,omitempty"`    // ug/m3
	CO2               float64 `json:"co2,omitempty"`     // ppm
	UVIndex           float64 `json:"uvIndex,omitempty"` // UV Index
	LightIntensity    float64 `json:"lightIntensity,omitempty"` // W/m2
	Pressure          float64 `json:"pressure"`    // hPa
	AirPressure       float64 `json:"airPressure"` // hPa alias
	Status            string  `json:"status"`       // "normal" | "warning" | "danger"
	OnlineStatus      string  `json:"onlineStatus"` // "online" | "offline"
	TimeOnsite        string  `json:"timeOnsite,omitempty"`
	LastUpdated       string  `json:"lastUpdated"`
	UpdatedAt         string  `json:"updatedAt"`
}

// WeatherDashboardStats represents summary statistics for Web Admin
type WeatherDashboardStats struct {
	TotalRegisters int    `json:"totalRegisters"`
	OnlineDevices  int    `json:"onlineDevices"`
	SilentDevices  int    `json:"silentDevices"`
	Malfunctions   int    `json:"malfunctions"`
	SyncTime       string `json:"syncTime"`
}

// WeatherTileFrame represents a single radar or forecast playback frame
type WeatherTileFrame struct {
	TimeLabel   string `json:"timeLabel"`
	DisplayTime string `json:"displayTime"`
	Formatted   string `json:"formatted"`
	URL         string `json:"url"`
	ValidTime   string `json:"valid_time,omitempty"`
}

// WeatherQueryFilter represents query params for filtering weather stations
type WeatherQueryFilter struct {
	Scope     string // "pattaya" or "all" (default: "all")
	AdminZone string
	District  string
	Search    string
	Status    string // "normal" | "warning" | "danger" | "all"
	Page      int
	PageSize  int
}

// WeatherAlertSetting represents real persisted alert rules in PostgreSQL
type WeatherAlertSetting struct {
	ID              string    `gorm:"primaryKey;column:id" json:"id"`
	Name            string    `gorm:"column:name;not null" json:"name"`
	Parameter       string    `gorm:"column:parameter;not null" json:"parameter"` // pm25, rain_hourly, rain_daily, etc.
	Condition       string    `gorm:"column:condition;not null" json:"condition"` // greater_than, less_than
	Value           float64   `gorm:"column:value;not null" json:"value"`
	Unit            string    `gorm:"column:unit;not null" json:"unit"`
	Severity        string    `gorm:"column:severity;not null" json:"severity"` // info, warning, danger
	MessageTitle    string    `gorm:"column:message_title;not null" json:"message_title"`
	MessageBody     string    `gorm:"column:message_body;not null" json:"message_body"`
	ImageMode       string    `gorm:"column:image_mode;default:'auto_card'" json:"image_mode"`
	CustomImageURL  string    `gorm:"column:custom_image_url" json:"custom_image_url,omitempty"`
	CooldownMinutes int       `gorm:"column:cooldown_minutes;default:60" json:"cooldown_minutes"`
	SendSMS         bool      `gorm:"column:send_sms;default:false" json:"send_sms"`
	SendPush        bool      `gorm:"column:send_push;default:true" json:"send_push"`
	Status          string    `gorm:"column:status;default:'online'" json:"status"`
	CreatedBy       string    `gorm:"column:created_by" json:"created_by"`
	CreatedDate     string    `gorm:"column:created_date" json:"created_date"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (WeatherAlertSetting) TableName() string {
	return "weather_alert_settings"
}

// WeatherCardTemplate represents background images for 3 alert severities
type WeatherCardTemplate struct {
	Severity      string    `gorm:"primaryKey;column:severity" json:"severity"` // info, warning, danger
	BackgroundURL string    `gorm:"column:background_url;not null" json:"background_url"`
	UpdatedBy     string    `gorm:"column:updated_by" json:"updated_by"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (WeatherCardTemplate) TableName() string {
	return "weather_card_templates"
}

// WeatherStationMetadata represents custom aliases, landmarks, and overrides in Chonburi DB
type WeatherStationMetadata struct {
	Serial       string    `gorm:"primaryKey;column:serial" json:"serial"`
	PoleCode     string    `gorm:"column:pole_code" json:"pole_code"`
	CustomName   string    `gorm:"column:custom_name;not null" json:"custom_name"`
	LandmarkDesc string    `gorm:"column:landmark_desc" json:"landmark_desc"`
	Subdistrict  string    `gorm:"column:subdistrict" json:"subdistrict"`
	District     string    `gorm:"column:district" json:"district"`
	AdminZone    string    `gorm:"column:admin_zone" json:"admin_zone"`
	UpdatedBy    string    `gorm:"column:updated_by" json:"updated_by"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (WeatherStationMetadata) TableName() string {
	return "weather_station_metadata"
}

// TelemetryHistoryRecord represents a historical IoT sensor telemetry record from TimescaleDB
type TelemetryHistoryRecord struct {
	Time           string  `json:"time"`
	TimeOnsite     string  `json:"time_onsite"`
	Temp           float64 `json:"temp"`
	Humid          float64 `json:"humid"`
	Press          float64 `json:"press"`
	RainRate       float64 `json:"rain_rate"`
	RainHourly     float64 `json:"rain_hourly"`
	RainDaily      float64 `json:"rain_daily"`
	PM1            float64 `json:"pm1"`
	PM25           float64 `json:"pm25"`
	PM40           float64 `json:"pm40"`
	PM10           float64 `json:"pm10"`
	WindSpeed      float64 `json:"wind_speed"`
	WindDirection  float64 `json:"wind_direction"`
	CO2            float64 `json:"co2"`
	UV             float64 `json:"uv"`
	Battery        float64 `json:"battery"`
	SignalStrength float64 `json:"signal_strength"`
}
