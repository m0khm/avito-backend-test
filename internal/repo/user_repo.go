package repo

import (
	"context"
	"fmt"
	"time"

	"room-booking-service/internal/domain"
	"room-booking-service/internal/tx"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Upsert(ctx context.Context, user domain.User) error {
	q := `
		INSERT INTO users (id, email, role, password_hash, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email, role = EXCLUDED.role`
	_, err := tx.DB(ctx, r.pool).Exec(ctx, q, user.ID, user.Email, user.Role, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}
	return nil
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (*domain.User, error) {
	q := `
		INSERT INTO users (id, email, role, password_hash, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at`
	var createdAt time.Time
	if err := tx.DB(ctx, r.pool).QueryRow(ctx, q, user.ID, user.Email, user.Role, user.PasswordHash).Scan(&createdAt); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	user.CreatedAt = &createdAt
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	q := `SELECT id, email, role, password_hash, created_at FROM users WHERE email = $1`
	var user domain.User
	var role string
	var passwordHash *string
	var createdAt time.Time
	if err := tx.DB(ctx, r.pool).QueryRow(ctx, q, email).Scan(&user.ID, &user.Email, &role, &passwordHash, &createdAt); err != nil {
		return nil, err
	}
	user.Role = domain.Role(role)
	user.PasswordHash = passwordHash
	user.CreatedAt = &createdAt
	return &user, nil
}
