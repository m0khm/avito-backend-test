package repo

import (
	"context"
	"fmt"
	"time"

	"room-booking-service/internal/domain"
	"room-booking-service/internal/tx"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepository struct {
	pool *pgxpool.Pool
}

func NewScheduleRepository(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{pool: pool}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule domain.Schedule) (*domain.Schedule, error) {
	days := make([]int32, 0, len(schedule.DaysOfWeek))
	for _, d := range schedule.DaysOfWeek {
		days = append(days, int32(d))
	}
	q := `
		INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING created_at`
	var createdAt time.Time
	if err := tx.DB(ctx, r.pool).QueryRow(ctx, q, schedule.ID, schedule.RoomID, days, schedule.StartTime, schedule.EndTime).Scan(&createdAt); err != nil {
		return nil, fmt.Errorf("create schedule: %w", err)
	}
	schedule.CreatedAt = &createdAt
	return &schedule, nil
}

func (r *ScheduleRepository) GetByRoomID(ctx context.Context, roomID string) (*domain.Schedule, error) {
	q := `SELECT id, room_id, days_of_week, start_time, end_time, created_at FROM schedules WHERE room_id = $1`
	var schedule domain.Schedule
	var days []int32
	var createdAt time.Time
	if err := tx.DB(ctx, r.pool).QueryRow(ctx, q, roomID).Scan(&schedule.ID, &schedule.RoomID, &days, &schedule.StartTime, &schedule.EndTime, &createdAt); err != nil {
		return nil, err
	}
	schedule.DaysOfWeek = make([]int, 0, len(days))
	for _, d := range days {
		schedule.DaysOfWeek = append(schedule.DaysOfWeek, int(d))
	}
	schedule.CreatedAt = &createdAt
	return &schedule, nil
}

func (r *ScheduleRepository) List(ctx context.Context) ([]domain.Schedule, error) {
	q := `SELECT id, room_id, days_of_week, start_time, end_time, created_at FROM schedules ORDER BY created_at`
	rows, err := tx.DB(ctx, r.pool).Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list schedules: %w", err)
	}
	defer rows.Close()

	var schedules []domain.Schedule
	for rows.Next() {
		var schedule domain.Schedule
		var days []int32
		var createdAt time.Time
		if err := rows.Scan(&schedule.ID, &schedule.RoomID, &days, &schedule.StartTime, &schedule.EndTime, &createdAt); err != nil {
			return nil, fmt.Errorf("scan schedule: %w", err)
		}
		schedule.DaysOfWeek = make([]int, 0, len(days))
		for _, d := range days {
			schedule.DaysOfWeek = append(schedule.DaysOfWeek, int(d))
		}
		schedule.CreatedAt = &createdAt
		schedules = append(schedules, schedule)
	}
	return schedules, rows.Err()
}
