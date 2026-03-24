package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"room-booking-service/internal/domain"
	"room-booking-service/internal/tx"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SlotRepository struct {
	pool *pgxpool.Pool
}

func NewSlotRepository(pool *pgxpool.Pool) *SlotRepository {
	return &SlotRepository{pool: pool}
}

func (r *SlotRepository) BulkUpsert(ctx context.Context, slots []domain.Slot) error {
	if len(slots) == 0 {
		return nil
	}

	var sb strings.Builder
	args := make([]any, 0, len(slots)*4)

	sb.WriteString(`
		INSERT INTO slots (id, room_id, start_at, end_at)
		VALUES
	`)

	for i, slot := range slots {
		if i > 0 {
			sb.WriteString(",")
		}

		base := i * 4
		sb.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4))
		args = append(args, slot.ID, slot.RoomID, slot.Start, slot.End)
	}

	sb.WriteString(`
		ON CONFLICT DO NOTHING
	`)

	if _, err := tx.DB(ctx, r.pool).Exec(ctx, sb.String(), args...); err != nil {
		return fmt.Errorf("bulk upsert slots: %w", err)
	}

	return nil
}

func (r *SlotRepository) ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]domain.Slot, error) {
	start, end := date, date.Add(24*time.Hour)
	q := `
		SELECT s.id, s.room_id, s.start_at, s.end_at
		FROM slots s
		LEFT JOIN bookings b
		  ON b.slot_id = s.id AND b.status = 'active'
		WHERE s.room_id = $1
		  AND s.start_at >= $2
		  AND s.start_at < $3
		  AND s.start_at >= NOW()
		  AND b.id IS NULL
		ORDER BY s.start_at`
	rows, err := tx.DB(ctx, r.pool).Query(ctx, q, roomID, start, end)
	if err != nil {
		return nil, fmt.Errorf("list available slots: %w", err)
	}
	defer rows.Close()

	var slots []domain.Slot
	for rows.Next() {
		var slot domain.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End); err != nil {
			return nil, fmt.Errorf("scan slot: %w", err)
		}
		slots = append(slots, slot)
	}

	return slots, rows.Err()
}

func (r *SlotRepository) GetByID(ctx context.Context, slotID string) (*domain.Slot, error) {
	q := `SELECT id, room_id, start_at, end_at FROM slots WHERE id = $1`
	var slot domain.Slot
	if err := tx.DB(ctx, r.pool).QueryRow(ctx, q, slotID).Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End); err != nil {
		return nil, err
	}
	return &slot, nil
}