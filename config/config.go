package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	AppPort            string
	AutoMigrate        bool
	DBDsn              string
	OrgName            string
	OrgLogoURL         string
	FrontendURL        string
	SMTPHost           string
	SMTPPort           string
	SMTPEmail          string
	SMTPPassword       string
	CORSAllowedOrigins string
	MinIO              MinIOConfig
	Redis              RedisConfig

	// Fahfon River Microservice (WireGuard: 10.1.2.19:8081)
	RiverServiceURL     string
	RiverGo2RTCURL      string
	RiverMunicipalityID string

	// Fahfon Weather Core Microservice (10.1.2.25:50059)
	FahfonWeatherBaseURL string
	FahfonWeatherMapURL  string
	FahfonServiceKey     string

	// M2M Dashboard Security
	M2MSecretKey string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type MinIOConfig struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	Bucket          string
	Region          string
	Secure          bool
	PublicRead      bool
	PublicBaseURL   string
	PresignURLTTL   int
	MaxUploadSizeMB int64
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Warning: No .env file found. Falling back to system environment variables.")
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("❌ FATAL: DB_DSN environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	orgName := os.Getenv("ORG_NAME")
	if orgName == "" {
		orgName = "องค์การบริหารส่วนจังหวัดชลบุรี"
	}

	orgLogoURL := os.Getenv("ORG_LOGO_URL")

	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "122.155.169.235:9000"
	}

	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioBucket := os.Getenv("MINIO_BUCKET")
	if minioBucket == "" {
		minioBucket = "super-app-assets"
	}

	minioRegion := os.Getenv("MINIO_REGION")
	if minioRegion == "" {
		minioRegion = "ap-southeast-1"
	}

	minioSecure := os.Getenv("MINIO_SECURE")
	useSecure := minioSecure == "true" || minioSecure == "1"

	minioPublicRead := getEnvBool("MINIO_PUBLIC_READ", false)
	minioPublicBaseURL := os.Getenv("MINIO_PUBLIC_BASE_URL")
	if minioPublicBaseURL == "" {
		scheme := "http"
		if useSecure {
			scheme = "https"
		}
		minioPublicBaseURL = scheme + "://" + minioEndpoint
	}

	minioPresignURLTTL := getEnvInt("MINIO_PRESIGN_URL_TTL_SECONDS", 3600)
	minioMaxUploadSizeMB := int64(getEnvInt("MINIO_MAX_UPLOAD_SIZE_MB", 10))

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587"
	}
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "127.0.0.1"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	corsAllowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsAllowedOrigins == "" {
		corsAllowedOrigins = "http://localhost:3000,http://localhost:3001,http://localhost:5173,https://uat-super-app-chonburi.mueangsmart.com"
	}

	riverServiceURL := os.Getenv("RIVER_SERVICE_URL")
	if riverServiceURL == "" {
		riverServiceURL = "http://10.1.2.19:8081"
	}
	riverGo2RTCURL := os.Getenv("RIVER_GO2RTC_URL")
	if riverGo2RTCURL == "" {
		riverGo2RTCURL = "http://10.1.2.7:1984"
	}
	riverMunicipalityID := os.Getenv("RIVER_MUNICIPALITY_ID")
	if riverMunicipalityID == "" {
		riverMunicipalityID = "c04da467-53f7-4076-93e9-54896eedb6c6"
	}
	fahfonWeatherBaseURL := os.Getenv("FAHFON_WEATHER_API_URL")
	if fahfonWeatherBaseURL == "" {
		fahfonWeatherBaseURL = os.Getenv("WEATHER_SERVICE_URL")
	}
	if fahfonWeatherBaseURL == "" {
		fahfonWeatherBaseURL = "http://10.1.2.25:50059"
	}
	fahfonWeatherMapURL := os.Getenv("FAHFON_WEATHER_MAP_URL")
	if fahfonWeatherMapURL == "" {
		fahfonWeatherMapURL = "http://10.1.2.25:50059"
	}
	fahfonServiceKey := os.Getenv("FAHFON_SERVICE_KEY")
	if fahfonServiceKey == "" {
		fahfonServiceKey = "ff_sec_service_key_Prod_7821_wQ9s8Zx2tA"
	}
	m2mSecretKey := os.Getenv("M2M_SECRET_KEY")
	if m2mSecretKey == "" {
		m2mSecretKey = "chonburi_m2m_sec_99a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4"
	}

	return &Config{
		AppEnv:               appEnv,
		AppPort:              port,
		AutoMigrate:          getEnvBool("AUTO_MIGRATE", true),
		DBDsn:                dsn,
		OrgName:              orgName,
		OrgLogoURL:           orgLogoURL,
		FrontendURL:          frontendURL,
		SMTPHost:             smtpHost,
		SMTPPort:             smtpPort,
		SMTPEmail:            smtpEmail,
		SMTPPassword:         smtpPassword,
		CORSAllowedOrigins:   corsAllowedOrigins,
		RiverServiceURL:      riverServiceURL,
		RiverGo2RTCURL:       riverGo2RTCURL,
		RiverMunicipalityID:  riverMunicipalityID,
		FahfonWeatherBaseURL: fahfonWeatherBaseURL,
		FahfonWeatherMapURL:  fahfonWeatherMapURL,
		FahfonServiceKey:     fahfonServiceKey,
		M2MSecretKey:         m2mSecretKey,
		MinIO: MinIOConfig{
			Endpoint:        minioEndpoint,
			AccessKey:       minioAccessKey,
			SecretKey:       minioSecretKey,
			Bucket:          minioBucket,
			Region:          minioRegion,
			Secure:          useSecure,
			PublicRead:      minioPublicRead,
			PublicBaseURL:   minioPublicBaseURL,
			PresignURLTTL:   minioPresignURLTTL,
			MaxUploadSizeMB: minioMaxUploadSizeMB,
		},
		Redis: RedisConfig{
			Host:     redisHost,
			Port:     redisPort,
			Password: redisPassword,
		},
	}
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value == "true" || value == "1"
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
