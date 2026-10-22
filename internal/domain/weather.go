package domain

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
	RainRate          float64 `json:"rainRate"`    // mm/h
	RainDaily         float64 `json:"rainDaily"`   // mm
	WindSpeed         float64 `json:"windSpeed"`   // m/s
	WindDirection     float64 `json:"windDirection"`
	WindDirectionText string  `json:"windDirectionText"`
	PM25              float64 `json:"pm25"`        // ug/m3
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
