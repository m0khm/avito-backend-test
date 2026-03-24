package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"room-booking-service/internal/domain"
	apperrors "room-booking-service/internal/errors"
	"room-booking-service/pkg/jwtutil"
	"room-booking-service/pkg/passwordutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type userRepo interface {
	Upsert(ctx context.Context, user domain.User) error
	Create(ctx context.Context, user domain.User) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type AuthService struct {
	users        userRepo
	jwtSecret    string
	dummyAdminID string
	dummyUserID  string
}

func NewAuthService(users userRepo, jwtSecret, dummyAdminID, dummyUserID string) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret, dummyAdminID: dummyAdminID, dummyUserID: dummyUserID}
}

func (s *AuthService) EnsureDummyUsers(ctx context.Context) error {
	admin := domain.User{ID: s.dummyAdminID, Email: "admin@example.com", Role: domain.RoleAdmin}
	user := domain.User{ID: s.dummyUserID, Email: "user@example.com", Role: domain.RoleUser}
	if err := s.users.Upsert(ctx, admin); err != nil {
		return err
	}
	return s.users.Upsert(ctx, user)
}

func (s *AuthService) DummyLogin(ctx context.Context, role string) (string, error) {
	role = strings.ToLower(strings.TrimSpace(role))
	if role != string(domain.RoleAdmin) && role != string(domain.RoleUser) {
		return "", apperrors.New(apperrors.ErrInvalidRequest.Code, "role must be admin or user", http.StatusBadRequest)
	}
	if err := s.EnsureDummyUsers(ctx); err != nil {
		return "", fmt.Errorf("ensure dummy users: %w", err)
	}
	userID := s.dummyUserID
	if role == string(domain.RoleAdmin) {
		userID = s.dummyAdminID
	}
	return jwtutil.IssueToken(s.jwtSecret, userID, role, 24*time.Hour)
}

func (s *AuthService) Register(ctx context.Context, email, password, role string) (*domain.User, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	role = strings.ToLower(strings.TrimSpace(role))
	if role != string(domain.RoleAdmin) && role != string(domain.RoleUser) {
		return nil, apperrors.New(apperrors.ErrInvalidRequest.Code, "role must be admin or user", http.StatusBadRequest)
	}
	hash, err := passwordutil.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user := domain.User{ID: uuid.NewString(), Email: email, Role: domain.Role(role), PasswordHash: &hash}
	created, err := s.users.Create(ctx, user)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return nil, apperrors.New(apperrors.ErrInvalidRequest.Code, "email already taken", http.StatusBadRequest)
		}
		return nil, err
	}
	return created, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return "", apperrors.New(apperrors.ErrUnauthorized.Code, "invalid credentials", http.StatusUnauthorized)
	}
	password = strings.TrimSpace(password)
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(apperrors.ErrUnauthorized.Code, "invalid credentials", http.StatusUnauthorized)
		}
		return "", err
	}
	if user.PasswordHash == nil || passwordutil.Compare(*user.PasswordHash, password) != nil {
		return "", apperrors.New(apperrors.ErrUnauthorized.Code, "invalid credentials", http.StatusUnauthorized)
	}
	return jwtutil.IssueToken(s.jwtSecret, user.ID, string(user.Role), 24*time.Hour)
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", apperrors.New(apperrors.ErrInvalidRequest.Code, "email is required", http.StatusBadRequest)
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || !strings.EqualFold(parsed.Address, email) {
		return "", apperrors.New(apperrors.ErrInvalidRequest.Code, "email must be valid", http.StatusBadRequest)
	}
	return email, nil
}

func validatePassword(password string) error {
	password = strings.TrimSpace(password)
	if len(password) < 6 {
		return apperrors.New(apperrors.ErrInvalidRequest.Code, "password must be at least 6 characters", http.StatusBadRequest)
	}
	return nil
}
