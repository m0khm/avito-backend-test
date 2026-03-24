package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"room-booking-service/internal/domain"
	apperrors "room-booking-service/internal/errors"
)

type bookingListRepo struct {
	listAllBookings []domain.Booking
	listAllTotal    int
	futureBookings  []domain.Booking
}

func (r *bookingListRepo) Create(ctx context.Context, booking domain.Booking) (*domain.Booking, error) {
	copy := booking
	return &copy, nil
}

func (r *bookingListRepo) UpdateConferenceLink(ctx context.Context, bookingID string, conferenceLink *string) error {
	return nil
}

func (r *bookingListRepo) ListAll(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	return r.listAllBookings, r.listAllTotal, nil
}

func (r *bookingListRepo) ListFutureByUser(ctx context.Context, userID string) ([]domain.Booking, error) {
	return r.futureBookings, nil
}

func (r *bookingListRepo) GetByID(ctx context.Context, bookingID string) (*domain.Booking, error) {
	return nil, nil
}

func (r *bookingListRepo) Cancel(ctx context.Context, bookingID string) (*domain.Booking, error) {
	return nil, nil
}

type updateLinkErrorBookingRepo struct {
	*fakeBookingRepo
	updateErr error
}

func (r *updateLinkErrorBookingRepo) UpdateConferenceLink(ctx context.Context, bookingID string, conferenceLink *string) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	return r.fakeBookingRepo.UpdateConferenceLink(ctx, bookingID, conferenceLink)
}

func TestBookingService_CreateAndCancel(t *testing.T) {
	bookings := newFakeBookingRepo()
	slots := newFakeSlotRepo()
	slot := domain.Slot{ID: "11111111-1111-1111-1111-111111111111", RoomID: "33333333-3333-3333-3333-333333333333", Start: time.Now().UTC().Add(2 * time.Hour), End: time.Now().UTC().Add(150 * time.Minute)}
	_ = slots.BulkUpsert(context.Background(), []domain.Slot{slot})
	svc := NewBookingService(bookings, slots, fakeTxManager{}, fakeConference{link: "https://meet.mock/123"})

	created, err := svc.Create(context.Background(), slot.ID, "22222222-2222-2222-2222-222222222222", true)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Status != domain.BookingStatusActive || created.ConferenceLink == nil {
		t.Fatalf("unexpected booking: %+v", created)
	}
	cancelled, err := svc.Cancel(context.Background(), created.ID, "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if cancelled.Status != domain.BookingStatusCancelled {
		t.Fatalf("expected cancelled status, got %s", cancelled.Status)
	}
	cancelledAgain, err := svc.Cancel(context.Background(), created.ID, "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("second Cancel() error = %v", err)
	}
	if cancelledAgain.Status != domain.BookingStatusCancelled {
		t.Fatalf("expected cancelled status on second cancel, got %s", cancelledAgain.Status)
	}
}

func TestBookingService_RejectPastSlot(t *testing.T) {
	bookings := newFakeBookingRepo()
	slots := newFakeSlotRepo()
	_ = slots.BulkUpsert(context.Background(), []domain.Slot{{ID: "44444444-4444-4444-4444-444444444444", RoomID: "33333333-3333-3333-3333-333333333333", Start: time.Now().UTC().Add(-2 * time.Hour), End: time.Now().UTC().Add(-90 * time.Minute)}})
	svc := NewBookingService(bookings, slots, fakeTxManager{}, fakeConference{})
	if _, err := svc.Create(context.Background(), "44444444-4444-4444-4444-444444444444", "22222222-2222-2222-2222-222222222222", false); err == nil {
		t.Fatal("expected error for past slot")
	}
}

func TestBookingService_Create_ConferenceLinkFailureDoesNotBreakBooking(t *testing.T) {
	bookings := newFakeBookingRepo()
	slots := newFakeSlotRepo()
	slot := domain.Slot{ID: "55555555-5555-5555-5555-555555555555", RoomID: "33333333-3333-3333-3333-333333333333", Start: time.Now().UTC().Add(2 * time.Hour), End: time.Now().UTC().Add(150 * time.Minute)}
	_ = slots.BulkUpsert(context.Background(), []domain.Slot{slot})
	svc := NewBookingService(bookings, slots, fakeTxManager{}, fakeConference{err: errors.New("conference down")})

	created, err := svc.Create(context.Background(), slot.ID, "22222222-2222-2222-2222-222222222222", true)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ConferenceLink != nil {
		t.Fatalf("expected nil conference link, got %+v", created.ConferenceLink)
	}
	if _, err := bookings.GetByID(context.Background(), created.ID); err != nil {
		t.Fatalf("booking must remain created when conference link fails: %v", err)
	}
}

func TestBookingService_Create_UpdateConferenceLinkFailureDoesNotBreakBooking(t *testing.T) {
	repo := &updateLinkErrorBookingRepo{fakeBookingRepo: newFakeBookingRepo(), updateErr: errors.New("write failed")}
	slots := newFakeSlotRepo()
	slot := domain.Slot{ID: "66666666-6666-6666-6666-666666666666", RoomID: "33333333-3333-3333-3333-333333333333", Start: time.Now().UTC().Add(2 * time.Hour), End: time.Now().UTC().Add(150 * time.Minute)}
	_ = slots.BulkUpsert(context.Background(), []domain.Slot{slot})
	svc := NewBookingService(repo, slots, fakeTxManager{}, fakeConference{link: "https://meet.mock/fail-save"})

	created, err := svc.Create(context.Background(), slot.ID, "22222222-2222-2222-2222-222222222222", true)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ConferenceLink != nil {
		t.Fatalf("expected nil conference link in response when persisting failed, got %+v", created.ConferenceLink)
	}
	persisted, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if persisted.ConferenceLink != nil {
		t.Fatalf("expected nil persisted conference link, got %+v", persisted.ConferenceLink)
	}
}

func TestBookingService_Cancel_Forbidden(t *testing.T) {
	bookings := newFakeBookingRepo()
	slots := newFakeSlotRepo()
	slot := domain.Slot{ID: "77777777-7777-7777-7777-777777777777", RoomID: "33333333-3333-3333-3333-333333333333", Start: time.Now().UTC().Add(2 * time.Hour), End: time.Now().UTC().Add(150 * time.Minute)}
	_ = slots.BulkUpsert(context.Background(), []domain.Slot{slot})
	svc := NewBookingService(bookings, slots, fakeTxManager{}, fakeConference{})

	created, err := svc.Create(context.Background(), slot.ID, "22222222-2222-2222-2222-222222222222", false)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = svc.Cancel(context.Background(), created.ID, "99999999-9999-9999-9999-999999999999")
	if err == nil {
		t.Fatal("expected forbidden error")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok || appErr.Code != apperrors.ErrForbidden.Code {
		t.Fatalf("expected forbidden AppError, got %T %v", err, err)
	}
}

func TestBookingService_ListAll(t *testing.T) {
	repo := &bookingListRepo{
		listAllBookings: []domain.Booking{
			{ID: "b1", SlotID: "11111111-1111-1111-1111-111111111111", UserID: "22222222-2222-2222-2222-222222222222", Status: domain.BookingStatusActive},
		},
		listAllTotal: 1,
	}
	svc := NewBookingService(repo, nil, nil, nil)

	bookings, pagination, err := svc.ListAll(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(bookings) != 1 {
		t.Fatalf("expected 1 booking, got %d", len(bookings))
	}
	if pagination.Total != 1 {
		t.Fatalf("expected total=1, got %d", pagination.Total)
	}
}

func TestBookingService_ListAll_InvalidPagination(t *testing.T) {
	svc := NewBookingService(&bookingListRepo{}, nil, nil, nil)

	if _, _, err := svc.ListAll(context.Background(), 0, 20); err == nil {
		t.Fatal("expected error for invalid page")
	}
	if _, _, err := svc.ListAll(context.Background(), 1, 101); err == nil {
		t.Fatal("expected error for invalid pageSize")
	}
}

func TestBookingService_ListFutureByUser(t *testing.T) {
	repo := &bookingListRepo{
		futureBookings: []domain.Booking{
			{ID: "b1", SlotID: "11111111-1111-1111-1111-111111111111", UserID: "22222222-2222-2222-2222-222222222222", Status: domain.BookingStatusActive},
		},
	}
	svc := NewBookingService(repo, nil, nil, nil)

	bookings, err := svc.ListFutureByUser(context.Background(), "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("ListFutureByUser() error = %v", err)
	}
	if len(bookings) != 1 {
		t.Fatalf("expected 1 booking, got %d", len(bookings))
	}
}
