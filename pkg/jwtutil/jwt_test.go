package jwtutil

import (
	"testing"
	"time"
)

func TestIssueAndParseToken(t *testing.T) {
	secret := "test-secret"

	token, err := IssueToken(secret, "user-1", "admin", time.Hour)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	if claims.UserID != "user-1" {
		t.Fatalf("expected user_id=user-1, got %s", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Fatalf("expected role=admin, got %s", claims.Role)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("expected subject=user-1, got %s", claims.Subject)
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	secret := "test-secret"

	if _, err := ParseToken(secret, "not-a-jwt"); err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}