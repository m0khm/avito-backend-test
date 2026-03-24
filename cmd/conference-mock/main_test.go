package main

import (
	"os"
	"testing"
)

func TestEnv_Default(t *testing.T) {
	key := "CONF_TEST_ENV_DEFAULT"
	_ = os.Unsetenv(key)

	got := env(key, "fallback")
	if got != "fallback" {
		t.Fatalf("expected fallback, got %s", got)
	}
}

func TestEnv_FromOS(t *testing.T) {
	key := "CONF_TEST_ENV_VALUE"
	t.Setenv(key, "value")

	got := env(key, "fallback")
	if got != "value" {
		t.Fatalf("expected value, got %s", got)
	}
}

func TestEnvFloat_Default(t *testing.T) {
	key := "CONF_TEST_FLOAT_DEFAULT"
	_ = os.Unsetenv(key)

	got := envFloat(key, 1.25)
	if got != 1.25 {
		t.Fatalf("expected 1.25, got %v", got)
	}
}

func TestEnvFloat_FromOS(t *testing.T) {
	key := "CONF_TEST_FLOAT_VALUE"
	t.Setenv(key, "2.5")

	got := envFloat(key, 1.25)
	if got != 2.5 {
		t.Fatalf("expected 2.5, got %v", got)
	}
}