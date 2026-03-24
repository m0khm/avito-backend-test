package service

import (
	"context"
	"testing"

	apperrors "room-booking-service/internal/errors"
	"room-booking-service/pkg/jwtutil"
)

func TestAuthService_DummyLogin(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewAuthService(repo, "secret", "admin-id", "user-id")
	token, err := svc.DummyLogin(context.Background(), "admin")
	if err != nil {
		t.Fatalf("DummyLogin() error = %v", err)
	}
	claims, err := jwtutil.ParseToken("secret", token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != "admin-id" || claims.Role != "admin" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewAuthService(repo, "secret", "admin-id", "user-id")
	_, err := svc.Register(context.Background(), "john@example.com", "password", "user")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	token, err := svc.Login(context.Background(), "john@example.com", "password")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	claims, err := jwtutil.ParseToken("secret", token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.Role != "user" {
		t.Fatalf("expected user role, got %s", claims.Role)
	}
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	svc := NewAuthService(newFakeUserRepo(), "secret", "admin-id", "user-id")

	_, err := svc.Register(context.Background(), "bad-email", "password", "user")
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok || appErr.Code != apperrors.ErrInvalidRequest.Code {
		t.Fatalf("expected invalid request AppError, got %T %v", err, err)
	}
}

func TestAuthService_Register_ShortPassword(t *testing.T) {
	svc := NewAuthService(newFakeUserRepo(), "secret", "admin-id", "user-id")

	_, err := svc.Register(context.Background(), "john@example.com", "123", "user")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthService_Login_InvalidEmailFormatReturnsUnauthorized(t *testing.T) {
	svc := NewAuthService(newFakeUserRepo(), "secret", "admin-id", "user-id")

	_, err := svc.Login(context.Background(), "bad-email", "password")
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok || appErr.Code != apperrors.ErrUnauthorized.Code {
		t.Fatalf("expected unauthorized AppError, got %T %v", err, err)
	}
}
