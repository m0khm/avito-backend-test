package repo

import (
	"context"
	"fmt"
	"time"

	"room-booking-service/internal/domain"
	"room-booking-service/internal/tx"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository struct {
	pool *pgxpool.Pool
}

func NewBookingRepository(pool *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{pool: pool}
}

func (r *BookingRepository) Create(ctx context.Context, booking domain.Booking) (*domain.Booking, error) {
	q := `
		INSERT INTO bookings (id, slot_id, user_id, status, conference_link, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING created_at`
	var createdAt time.Time
	if err := tx.DB(ctx, r.pool).QueryRow(ctx, q, booking.ID, booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink).Scan(&createdAt); err != nil {
		return nil, fmt.Errorf("create booking: %w", err)
	}
	booking.CreatedAt = &createdAt
	return &booking, nil
}


func (r *BookingRepository) UpdateConferenceLink(ctx context.Context, bookingID string, conferenceLink *string) error {
	q := `
		UPDATE bookings
		SET conference_link = $2
		WHERE id = $1`
	if _, err := tx.DB(ctx, r.pool).Exec(ctx, q, bookingID, conferenceLink); err != nil {
		return fmt.Errorf("update booking conference link: %w", err)
	}
	return nil
}

func (r *BookingRepository) ListAll(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	countQ := `SELECT COUNT(*) FROM bookings`
	var total int
	if err := tx.DB(ctx, r.pool).QueryRow(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count bookings: %w", err)
	}
	q := `
		SELECT b.id, b.slot_id, b.user_id, b.status, b.conference_link, b.created_at
		FROM bookings b
		ORDER BY b.created_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := tx.DB(ctx, r.pool).Query(ctx, q, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()
	bookings := make([]domain.Booking, 0, pageSize)
	for rows.Next() {
		var booking domain.Booking
		var status string
		var createdAt time.Time
		if err := rows.Scan(&booking.ID, &booking.SlotID, &booking.UserID, &status, &booking.ConferenceLink, &createdAt); err != nil {
			return nil, 0, fmt.Errorf("scan booking: %w", err)
		}
		booking.Status = domain.BookingStatus(status)
		booking.CreatedAt = &createdAt
		bookings = append(bookings, booking)
	}
	return bookings, total, rows.Err()
}

func (r *BookingRepository) ListFutureByUser(ctx context.Context, userID string) ([]domain.Booking, error) {
	q := `
		SELECT b.id, b.slot_id, b.user_id, b.status, b.conference_link, b.created_at, s.start_at
		FROM bookings b
		JOIN slots s ON s.id = b.slot_id
		WHERE b.user_id = $1 AND b.status = 'active' AND s.start_at >= NOW()
		ORDER BY s.start_at, b.created_at`
	rows, err := tx.DB(ctx, r.pool).Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list future bookings: %w", err)
	}
	defer rows.Close()
	var bookings []domain.Booking
	for rows.Next() {
		var booking domain.Booking
		var status string
		var createdAt, slotStart time.Time
		if err := rows.Scan(&booking.ID, &booking.SlotID, &booking.UserID, &status, &booking.ConferenceLink, &createdAt, &slotStart); err != nil {
			return nil, fmt.Errorf("scan future booking: %w", err)
		}
		booking.Status = domain.BookingStatus(status)
		booking.CreatedAt = &createdAt
		booking.SlotStart = &slotStart
		bookings = append(bookings, booking)
	}
	return bookings, rows.Err()
}

func (r *BookingRepository) GetByID(ctx context.Context, bookingID string) (*domain.Booking, error) {
	q := `SELECT id, slot_id, user_id, status, conference_link, created_at FROM bookings WHERE id = $1`
	var booking domain.Booking
	var status string
	var createdAt time.Time
	if err := tx.DB(ctx, r.pool).QueryRow(ctx, q, bookingID).Scan(&booking.ID, &booking.SlotID, &booking.UserID, &status, &booking.ConferenceLink, &createdAt); err != nil {
		return nil, err
	}
	booking.Status = domain.BookingStatus(status)
	booking.CreatedAt = &createdAt
	return &booking, nil
}

func (r *BookingRepository) Cancel(ctx context.Context, bookingID string) (*domain.Booking, error) {
	q := `
		UPDATE bookings
		SET status = 'cancelled'
		WHERE id = $1
		RETURNING id, slot_id, user_id, status, conference_link, created_at`
	var booking domain.Booking
	var status string
	var createdAt time.Time
	if err := tx.DB(ctx, r.pool).QueryRow(ctx, q, bookingID).Scan(&booking.ID, &booking.SlotID, &booking.UserID, &status, &booking.ConferenceLink, &createdAt); err != nil {
		return nil, err
	}
	booking.Status = domain.BookingStatus(status)
	booking.CreatedAt = &createdAt
	return &booking, nil
}
