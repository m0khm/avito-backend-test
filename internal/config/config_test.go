package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("HTTP_PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("SLOT_HORIZON_DAYS", "")
	t.Setenv("CONFERENCE_SERVICE_URL", "")
	t.Setenv("CONFERENCE_TIMEOUT_MS", "")
	t.Setenv("DUMMY_ADMIN_USER_ID", "")
	t.Setenv("DUMMY_USER_USER_ID", "")
	t.Setenv("CONFERENCE_MOCK_PORT", "")

	cfg := Load()

	if cfg.HTTPPort != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.HTTPPort)
	}
	if cfg.JWTSecret == "" {
		t.Fatal("expected default JWT secret")
	}
	if cfg.SlotHorizonDays != 7 {
		t.Fatalf("expected default horizon 7, got %d", cfg.SlotHorizonDays)
	}
	if cfg.ConferenceTimeout != 1500*time.Millisecond {
		t.Fatalf("unexpected conference timeout: %v", cfg.ConferenceTimeout)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("JWT_SECRET", "jwt")
	t.Setenv("SLOT_HORIZON_DAYS", "7")
	t.Setenv("CONFERENCE_SERVICE_URL", "http://svc")
	t.Setenv("CONFERENCE_TIMEOUT_MS", "2500")
	t.Setenv("DUMMY_ADMIN_USER_ID", "admin-id")
	t.Setenv("DUMMY_USER_USER_ID", "user-id")
	t.Setenv("CONFERENCE_MOCK_PORT", "9000")

	cfg := Load()

	if cfg.HTTPPort != "9090" {
		t.Fatalf("expected 9090, got %s", cfg.HTTPPort)
	}
	if cfg.DatabaseURL != "postgres://x" {
		t.Fatalf("unexpected database url: %s", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "jwt" {
		t.Fatalf("unexpected jwt secret: %s", cfg.JWTSecret)
	}
	if cfg.SlotHorizonDays != 7 {
		t.Fatalf("unexpected horizon: %d", cfg.SlotHorizonDays)
	}
	if cfg.ConferenceServiceURL != "http://svc" {
		t.Fatalf("unexpected conference url: %s", cfg.ConferenceServiceURL)
	}
	if cfg.ConferenceTimeout != 2500*time.Millisecond {
		t.Fatalf("unexpected timeout: %v", cfg.ConferenceTimeout)
	}
	if cfg.DummyAdminUserID != "admin-id" || cfg.DummyUserUserID != "user-id" {
		t.Fatal("unexpected dummy ids")
	}
	if cfg.ConferenceMockPort != "9000" {
		t.Fatalf("unexpected conference mock port: %s", cfg.ConferenceMockPort)
	}
}

func TestGetEnv(t *testing.T) {
	key := "TEST_GET_ENV_VALUE"
	_ = os.Unsetenv(key)

	if got := getEnv(key, "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %s", got)
	}

	t.Setenv(key, "value")
	if got := getEnv(key, "fallback"); got != "value" {
		t.Fatalf("expected value, got %s", got)
	}
}