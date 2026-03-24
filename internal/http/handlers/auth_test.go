package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"room-booking-service/internal/domain"
	"room-booking-service/internal/service"
	"room-booking-service/pkg/passwordutil"

	"github.com/jackc/pgx/v5"
)

type authTestRepo struct {
	usersByEmail map[string]domain.User
}

func (r *authTestRepo) Upsert(ctx context.Context, user domain.User) error {
	if r.usersByEmail == nil {
		r.usersByEmail = map[string]domain.User{}
	}
	r.usersByEmail[user.Email] = user
	return nil
}

func (r *authTestRepo) Create(ctx context.Context, user domain.User) (*domain.User, error) {
	if r.usersByEmail == nil {
		r.usersByEmail = map[string]domain.User{}
	}
	r.usersByEmail[user.Email] = user
	u := user
	return &u, nil
}

func (r *authTestRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, ok := r.usersByEmail[email]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	copy := u
	return &copy, nil
}

func newAuthHandlerForTests(repo *authTestRepo) *Handler {
	authSvc := service.NewAuthService(
		repo,
		"secret",
		"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
	)
	return &Handler{auth: authSvc}
}

func TestDummyLogin(t *testing.T) {
	h := newAuthHandlerForTests(&authTestRepo{})

	body := []byte(`{"role":"admin"}`)
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.DummyLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestDummyLogin_InvalidBody(t *testing.T) {
	h := newAuthHandlerForTests(&authTestRepo{})

	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewReader([]byte(`{bad json}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.DummyLogin(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegister_Success(t *testing.T) {
	h := newAuthHandlerForTests(&authTestRepo{})

	body := []byte(`{"email":"user@example.com","password":"pass123","role":"user"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"user@example.com"`) {
		t.Fatalf("expected created user in response, body=%s", rec.Body.String())
	}
}

func TestRegister_InvalidBody(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte(`{bad json}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_Success(t *testing.T) {
	hash, err := passwordutil.Hash("pass123")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	h := newAuthHandlerForTests(&authTestRepo{
		usersByEmail: map[string]domain.User{
			"user@example.com": {
				ID:           "user-1",
				Email:        "user@example.com",
				Role:         domain.RoleUser,
				PasswordHash: &hash,
			},
		},
	})

	body := []byte(`{"email":"user@example.com","password":"pass123"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"token"`) {
		t.Fatalf("expected token in response, body=%s", rec.Body.String())
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	h := newAuthHandlerForTests(&authTestRepo{})

	body := []byte(`{"email":"missing@example.com","password":"pass123"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d, body=%s", rec.Code, rec.Body.String())
	}
}
