package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"room-booking-service/internal/domain"
	apperrors "room-booking-service/internal/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type bookingRepo interface {
	Create(ctx context.Context, booking domain.Booking) (*domain.Booking, error)
	UpdateConferenceLink(ctx context.Context, bookingID string, conferenceLink *string) error
	ListAll(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error)
	ListFutureByUser(ctx context.Context, userID string) ([]domain.Booking, error)
	GetByID(ctx context.Context, bookingID string) (*domain.Booking, error)
	Cancel(ctx context.Context, bookingID string) (*domain.Booking, error)
}

type txManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type conferenceCreator interface {
	CreateConferenceLink(ctx context.Context, bookingID, slotID, userID string) (*string, error)
}

type BookingService struct {
	bookings   bookingRepo
	slots      slotRepo
	txManager  txManager
	conference conferenceCreator
}

func NewBookingService(bookings bookingRepo, slots slotRepo, txManager txManager, conference conferenceCreator) *BookingService {
	return &BookingService{bookings: bookings, slots: slots, txManager: txManager, conference: conference}
}

func (s *BookingService) Create(ctx context.Context, slotID, userID string, createConferenceLink bool) (*domain.Booking, error) {
	if err := validateUUID(slotID, "slotId"); err != nil {
		return nil, err
	}

	bookingID := uuid.NewString()
	var created *domain.Booking

	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		slot, err := s.slots.GetByID(txCtx, slotID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperrors.ErrSlotNotFound
			}
			if isInvalidUUIDError(err) {
				return apperrors.ErrInvalidRequest
			}
			return err
		}

		if slot.Start.Before(time.Now().UTC()) {
			return apperrors.New(apperrors.ErrInvalidRequest.Code, "cannot create booking for a past slot", http.StatusBadRequest)
		}

		booking := domain.Booking{
			ID:     bookingID,
			SlotID: slotID,
			UserID: userID,
			Status: domain.BookingStatusActive,
		}

		createdBooking, createErr := s.bookings.Create(txCtx, booking)
		if createErr != nil {
			if isUniqueViolation(createErr) {
				return apperrors.ErrSlotAlreadyBooked
			}
			if isInvalidUUIDError(createErr) {
				return apperrors.ErrInvalidRequest
			}
			return createErr
		}

		created = createdBooking
		return nil
	})
	if err != nil {
		return nil, err
	}

	if createConferenceLink {
		s.attachConferenceLinkBestEffort(ctx, created, userID)
	}

	return created, nil
}

func (s *BookingService) attachConferenceLinkBestEffort(ctx context.Context, booking *domain.Booking, userID string) {
	if booking == nil || s.conference == nil {
		return
	}

	link, err := s.conference.CreateConferenceLink(ctx, booking.ID, booking.SlotID, userID)
	if err != nil || link == nil || *link == "" {
		return
	}

	if err := s.bookings.UpdateConferenceLink(ctx, booking.ID, link); err != nil {
		return
	}

	booking.ConferenceLink = link
}

func (s *BookingService) ListAll(ctx context.Context, page, pageSize int) ([]domain.Booking, domain.Pagination, error) {
	if page < 1 || pageSize < 1 || pageSize > 100 {
		return nil, domain.Pagination{}, apperrors.New(apperrors.ErrInvalidRequest.Code, "invalid pagination params", http.StatusBadRequest)
	}

	bookings, total, err := s.bookings.ListAll(ctx, page, pageSize)
	if err != nil {
		return nil, domain.Pagination{}, err
	}

	return bookings, domain.Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *BookingService) ListFutureByUser(ctx context.Context, userID string) ([]domain.Booking, error) {
	return s.bookings.ListFutureByUser(ctx, userID)
}

func (s *BookingService) Cancel(ctx context.Context, bookingID, userID string) (*domain.Booking, error) {
	booking, err := s.bookings.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrBookingNotFound
		}
		if isInvalidUUIDError(err) {
			return nil, apperrors.ErrInvalidRequest
		}
		return nil, err
	}

	if booking.UserID != userID {
		return nil, apperrors.New(apperrors.ErrForbidden.Code, "cannot cancel another user's booking", http.StatusForbidden)
	}

	if booking.Status == domain.BookingStatusCancelled {
		return booking, nil
	}

	cancelled, err := s.bookings.Cancel(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("cancel booking: %w", err)
	}

	return cancelled, nil
}
