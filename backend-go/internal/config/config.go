package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                     string
	Env                      string
	DBDriver                 string // "postgres" or "sqlite"
	DBDSN                    string
	JWTSecret                string
	JWTRefreshSecret         string
	JWTAccessExpiryHours     int
	JWTRefreshExpiryDays     int
	AppleSharedSecret        string
	AndroidPackageName       string
	GooglePlayServiceAccount string
	APNSKeyID                string
	APNSTeamID               string
	APNSBundleID             string
	APNSKeyPath              string
	APNSUseSandbox           bool
	AgoraAppID               string
	AgoraAppCertificate      string
	StripeSecretKey          string
	StripePublishableKey     string
	StripeWebhookSecret      string
	CheckrAPIKey             string
	FrontendURL              string
	MediaUploadDir           string
	FirebaseCredentialsFile  string
	FirebaseCredentialsJSON  string
	SMTPHost                 string
	SMTPPort                 int
	SMTPUser                 string
	SMTPPassword             string
	EmailFrom                string
	AdminEmail               string
	RedisAddr                string
	RedisPassword            string
	RedisDB                  int
	RedisURL                 string
}

func Load() *Config {
	_ = godotenv.Load(".env", "../backend/.env", "../.env")

	port := getEnv("PORT", "8000")
	env := getEnv("ENV", "development")
	dbDriver := getEnv("DB_DRIVER", "postgres")

	pgHost := getEnv("POSTGRES_HOST", "localhost")
	pgPort := getEnv("POSTGRES_PORT", "5432")
	pgUser := getEnv("POSTGRES_USER", "postgres")
	pgPass := getEnv("POSTGRES_PASSWORD", "gentechco")
	pgName := getEnv("POSTGRES_DB", "nsapp")
	defaultDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", pgHost, pgPort, pgUser, pgPass, pgName)

	dbDSN := getEnv("DATABASE_URL", getEnv("DB_DSN", defaultDSN))
	jwtSecret := getEnv("JWT_SECRET", getEnv("SECRET_KEY", "django-insecure-ns-secret-key-change-in-production"))
	jwtRefreshSecret := getEnv("JWT_REFRESH_SECRET", jwtSecret+"-refresh")
	
	accessExpiry, _ := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRY_HOURS", "24"))
	if accessExpiry <= 0 {
		accessExpiry = 24
	}
	refreshExpiry, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRY_DAYS", "30"))
	if refreshExpiry <= 0 {
		refreshExpiry = 30
	}

	smtpPort, _ := strconv.Atoi(getEnv("EMAIL_PORT", "587"))
	if smtpPort <= 0 {
		smtpPort = 587
	}

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisAddr := getEnv("REDIS_ADDR", fmt.Sprintf("%s:%s", redisHost, redisPort))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		Port:                     port,
		Env:                      env,
		DBDriver:                 dbDriver,
		DBDSN:                    dbDSN,
		JWTSecret:                jwtSecret,
		JWTRefreshSecret:         jwtRefreshSecret,
		JWTAccessExpiryHours:     accessExpiry,
		JWTRefreshExpiryDays:     refreshExpiry,
		AppleSharedSecret:        getEnv("APPLE_SHARED_SECRET", ""),
		AndroidPackageName:       getEnv("ANDROID_PACKAGE_NAME", "com.neighborservicesolutionsllc.nsapp"),
		GooglePlayServiceAccount: getEnv("GOOGLE_PLAY_SERVICE_ACCOUNT_JSON", ""),
		APNSKeyID:                getEnv("APNS_KEY_ID", ""),
		APNSTeamID:               getEnv("APNS_TEAM_ID", ""),
		APNSBundleID:             getEnv("APNS_BUNDLE_ID", "com.neighborservicesolutionsllc.nsapp"),
		APNSKeyPath:              getEnv("APNS_KEY_PATH", "./apns.p8"),
		APNSUseSandbox:           getEnv("APNS_USE_SANDBOX", "True") == "True",
		AgoraAppID:               getEnv("AGORA_APP_ID", ""),
		AgoraAppCertificate:      getEnv("AGORA_APP_CERTIFICATE", ""),
		StripeSecretKey:          getEnv("STRIPE_SECRET_KEY", ""),
		StripePublishableKey:     getEnv("STRIPE_PUBLIC_KEY", getEnv("STRIPE_PUBLISHABLE_KEY", "")),
		StripeWebhookSecret:      getEnv("STRIPE_WEBHOOK_SECRET", ""),
		CheckrAPIKey:             getEnv("CHECKR_API_KEY", ""),
		FrontendURL:              getEnv("FRONTEND_URL", "http://localhost:4200"),
		MediaUploadDir:           getEnv("MEDIA_UPLOAD_DIR", "./media"),
		FirebaseCredentialsFile:  getEnv("FIREBASE_CREDENTIALS_FILE", "./firebase-service-account.json"),
		FirebaseCredentialsJSON:  getEnv("FIREBASE_CREDENTIALS_JSON", ""),
		SMTPHost:                 getEnv("EMAIL_HOST", ""),
		SMTPPort:                 smtpPort,
		SMTPUser:                 getEnv("EMAIL_HOST_USER", ""),
		SMTPPassword:             getEnv("EMAIL_HOST_PASSWORD", ""),
		EmailFrom:                getEnv("DEFAULT_FROM_EMAIL", getEnv("EMAIL_HOST_USER", "noreply@neighborservice.com")),
		AdminEmail:               getEnv("ADMIN_EMAIL", getEnv("DEFAULT_FROM_EMAIL", getEnv("EMAIL_HOST_USER", "admin@neighborservice.com"))),
		RedisAddr:                redisAddr,
		RedisPassword:            getEnv("REDIS_PASSWORD", ""),
		RedisDB:                  redisDB,
		RedisURL:                 getEnv("REDIS_URL", ""),
	}
}


func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}
