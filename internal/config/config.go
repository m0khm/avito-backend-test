package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort             string
	DatabaseURL          string
	JWTSecret            string
	SlotHorizonDays      int
	ConferenceServiceURL string
	ConferenceTimeout    time.Duration
	DummyAdminUserID     string
	DummyUserUserID      string
	ConferenceMockPort   string
}

func Load() Config {
	return Config{
		HTTPPort:             getEnv("HTTP_PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@db:5432/rooms?sslmode=disable"),
		JWTSecret:            getEnv("JWT_SECRET", "super-secret-key"),
		SlotHorizonDays:      getEnvInt("SLOT_HORIZON_DAYS", 7),
		ConferenceServiceURL: getEnv("CONFERENCE_SERVICE_URL", "http://conference-mock:8090"),
		ConferenceTimeout:    time.Duration(getEnvInt("CONFERENCE_TIMEOUT_MS", 1500)) * time.Millisecond,
		DummyAdminUserID:     getEnv("DUMMY_ADMIN_USER_ID", "11111111-1111-1111-1111-111111111111"),
		DummyUserUserID:      getEnv("DUMMY_USER_USER_ID", "22222222-2222-2222-2222-222222222222"),
		ConferenceMockPort:   getEnv("CONFERENCE_MOCK_PORT", "8090"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
		panic(fmt.Sprintf("invalid integer env %s", key))
	}
	return fallback
}
