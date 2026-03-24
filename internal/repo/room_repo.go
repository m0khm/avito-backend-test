package repo

import (
	"context"
	"fmt"
	"time"

	"room-booking-service/internal/domain"
	"room-booking-service/internal/tx"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	pool *pgxpool.Pool
}

func NewRoomRepository(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

func (r *RoomRepository) Create(ctx context.Context, room domain.Room) (*domain.Room, error) {
	q := `
		INSERT INTO rooms (id, name, description, capacity, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at`
	var createdAt time.Time
	if err := tx.DB(ctx, r.pool).QueryRow(ctx, q, room.ID, room.Name, room.Description, room.Capacity).Scan(&createdAt); err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	room.CreatedAt = &createdAt
	return &room, nil
}

func (r *RoomRepository) List(ctx context.Context) ([]domain.Room, error) {
	q := `SELECT id, name, description, capacity, created_at FROM rooms ORDER BY created_at, name`
	rows, err := tx.DB(ctx, r.pool).Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	var rooms []domain.Room
	for rows.Next() {
		var room domain.Room
		var createdAt time.Time
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &createdAt); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		room.CreatedAt = &createdAt
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}

func (r *RoomRepository) Exists(ctx context.Context, roomID string) (bool, error) {
	q := `SELECT EXISTS(SELECT 1 FROM rooms WHERE id = $1)`
	var exists bool
	if err := tx.DB(ctx, r.pool).QueryRow(ctx, q, roomID).Scan(&exists); err != nil {
		return false, fmt.Errorf("room exists: %w", err)
	}
	return exists, nil
}
