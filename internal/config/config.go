package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort          string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBSchema            string
	JWTSecret           string
	JWTExpirationMs     int64
	PortfolioServiceURL string
}

func Load() *Config {
	return &Config{
		ServerPort:          getEnv("SERVER_PORT", "8081"),
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "5432"),
		DBUser:              getEnv("DB_USER", "auth_svc_user"),
		DBPassword:          getEnv("DB_PASSWORD", "localdev123"),
		DBName:              getEnv("DB_NAME", "wmp"),
		DBSchema:            getEnv("DB_SCHEMA", "auth_schema"),
		JWTSecret:           getEnv("JWT_SECRET", "myDefaultSuperSecretKeyThatIsAtLeast256BitsLongForHS256Algorithm"),
		JWTExpirationMs:     getEnvInt64("JWT_EXPIRATION_MS", 86400000),
		PortfolioServiceURL: getEnv("PORTFOLIO_SERVICE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}
