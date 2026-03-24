package repo

import (
	"context"
	"testing"
	"time"

	"room-booking-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

func createBookingDeps(t *testing.T, db *pgxpool.Pool, roomID, roomName, slotID, userID, email string, start time.Time) (domain.Slot, domain.User, *BookingRepository) {
	t.Helper()
	roomRepo := NewRoomRepository(db)
	slotRepo := NewSlotRepository(db)
	userRepo := NewUserRepository(db)
	bookingRepo := NewBookingRepository(db)

	room := domain.Room{ID: roomID, Name: roomName}
	if _, err := roomRepo.Create(context.Background(), room); err != nil {
		t.Fatalf("Create room error = %v", err)
	}

	slot := domain.Slot{ID: slotID, RoomID: room.ID, Start: start, End: start.Add(30 * time.Minute)}
	if err := slotRepo.BulkUpsert(context.Background(), []domain.Slot{slot}); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	userCreatedAt := time.Now().UTC()
	user := domain.User{ID: userID, Email: email, Role: domain.RoleUser, CreatedAt: &userCreatedAt}
	if err := userRepo.Upsert(context.Background(), user); err != nil {
		t.Fatalf("Upsert user error = %v", err)
	}

	return slot, user, bookingRepo
}

func TestBookingRepository_CreateAndGetByID(t *testing.T) {
	db := testDB(t)
	slot, user, bookingRepo := createBookingDeps(
		t, db,
		"00000000-0000-0000-0000-000000000101", "Booking Room",
		"00000000-0000-0000-0000-000000000102",
		"00000000-0000-0000-0000-000000000104", "booking-test@example.com",
		time.Now().UTC().Add(2*time.Hour).Truncate(30*time.Minute),
	)

	bookingCreatedAt := time.Now().UTC()
	booking := domain.Booking{ID: "00000000-0000-0000-0000-000000000103", SlotID: slot.ID, UserID: user.ID, Status: domain.BookingStatusActive, CreatedAt: &bookingCreatedAt}
	created, err := bookingRepo.Create(context.Background(), booking)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID != booking.ID || created.Status != domain.BookingStatusActive {
		t.Fatalf("unexpected created booking: %+v", created)
	}

	got, err := bookingRepo.GetByID(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != booking.ID || got.SlotID != booking.SlotID || got.UserID != booking.UserID {
		t.Fatalf("unexpected booking: %+v", got)
	}
}

func TestBookingRepository_Cancel(t *testing.T) {
	db := testDB(t)
	slot, user, bookingRepo := createBookingDeps(
		t, db,
		"00000000-0000-0000-0000-000000000111", "Cancel Room",
		"00000000-0000-0000-0000-000000000112",
		"00000000-0000-0000-0000-000000000114", "cancel-test@example.com",
		time.Now().UTC().Add(3*time.Hour).Truncate(30*time.Minute),
	)

	booking := domain.Booking{ID: "00000000-0000-0000-0000-000000000113", SlotID: slot.ID, UserID: user.ID, Status: domain.BookingStatusActive}
	if _, err := bookingRepo.Create(context.Background(), booking); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	cancelled, err := bookingRepo.Cancel(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if cancelled.Status != domain.BookingStatusCancelled {
		t.Fatalf("expected cancelled, got %s", cancelled.Status)
	}
}

func TestBookingRepository_ListFutureByUser_FiltersPastAndCancelled(t *testing.T) {
	db := testDB(t)
	userCreatedAt := time.Now().UTC()
	userRepo := NewUserRepository(db)
	user := domain.User{ID: "00000000-0000-0000-0000-000000000300", Email: "future@example.com", Role: domain.RoleUser, CreatedAt: &userCreatedAt}
	if err := userRepo.Upsert(context.Background(), user); err != nil {
		t.Fatalf("Upsert user error = %v", err)
	}
	roomRepo := NewRoomRepository(db)
	slotRepo := NewSlotRepository(db)
	bookingRepo := NewBookingRepository(db)

	rooms := []domain.Room{
		{ID: "00000000-0000-0000-0000-000000000301", Name: "Future Room"},
		{ID: "00000000-0000-0000-0000-000000000302", Name: "Past Room"},
		{ID: "00000000-0000-0000-0000-000000000303", Name: "Cancelled Room"},
	}
	for _, room := range rooms {
		if _, err := roomRepo.Create(context.Background(), room); err != nil {
			t.Fatalf("Create room error = %v", err)
		}
	}
	futureStart := time.Now().UTC().Add(4 * time.Hour).Truncate(30 * time.Minute)
	pastStart := time.Now().UTC().Add(-4 * time.Hour).Truncate(30 * time.Minute)
	cancelledStart := time.Now().UTC().Add(5 * time.Hour).Truncate(30 * time.Minute)
	slots := []domain.Slot{
		{ID: "00000000-0000-0000-0000-000000000311", RoomID: rooms[0].ID, Start: futureStart, End: futureStart.Add(30 * time.Minute)},
		{ID: "00000000-0000-0000-0000-000000000312", RoomID: rooms[1].ID, Start: pastStart, End: pastStart.Add(30 * time.Minute)},
		{ID: "00000000-0000-0000-0000-000000000313", RoomID: rooms[2].ID, Start: cancelledStart, End: cancelledStart.Add(30 * time.Minute)},
	}
	if err := slotRepo.BulkUpsert(context.Background(), slots); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}
	for _, booking := range []domain.Booking{
		{ID: "00000000-0000-0000-0000-000000000321", SlotID: slots[0].ID, UserID: user.ID, Status: domain.BookingStatusActive},
		{ID: "00000000-0000-0000-0000-000000000322", SlotID: slots[1].ID, UserID: user.ID, Status: domain.BookingStatusActive},
		{ID: "00000000-0000-0000-0000-000000000323", SlotID: slots[2].ID, UserID: user.ID, Status: domain.BookingStatusCancelled},
	} {
		if _, err := bookingRepo.Create(context.Background(), booking); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	items, err := bookingRepo.ListFutureByUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("ListFutureByUser() error = %v", err)
	}
	if len(items) != 1 || items[0].SlotID != slots[0].ID {
		t.Fatalf("unexpected future bookings: %+v", items)
	}
}

func TestBookingRepository_UpdateConferenceLink(t *testing.T) {
	db := testDB(t)
	slot, user, bookingRepo := createBookingDeps(
		t, db,
		"00000000-0000-0000-0000-000000000601", "Conference Room",
		"00000000-0000-0000-0000-000000000602",
		"00000000-0000-0000-0000-000000000603", "conference-test@example.com",
		time.Now().UTC().Add(6*time.Hour).Truncate(30*time.Minute),
	)
	booking := domain.Booking{ID: "00000000-0000-0000-0000-000000000604", SlotID: slot.ID, UserID: user.ID, Status: domain.BookingStatusActive}
	if _, err := bookingRepo.Create(context.Background(), booking); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	link := "https://meet.example/booking-601"
	if err := bookingRepo.UpdateConferenceLink(context.Background(), booking.ID, &link); err != nil {
		t.Fatalf("UpdateConferenceLink() error = %v", err)
	}
	got, err := bookingRepo.GetByID(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ConferenceLink == nil || *got.ConferenceLink != link {
		t.Fatalf("expected conference link %q, got %+v", link, got.ConferenceLink)
	}
}

func TestBookingRepository_ListAll(t *testing.T) {
	db := testDB(t)

	roomRepo := NewRoomRepository(db)
	slotRepo := NewSlotRepository(db)
	userRepo := NewUserRepository(db)
	bookingRepo := NewBookingRepository(db)

	room := domain.Room{
		ID:   "00000000-0000-0000-0000-000000000701",
		Name: "ListAll Room",
	}
	if _, err := roomRepo.Create(context.Background(), room); err != nil {
		t.Fatalf("Create room error = %v", err)
	}

	start1 := time.Now().UTC().Add(2 * time.Hour).Truncate(30 * time.Minute)
	start2 := start1.Add(30 * time.Minute)

	slots := []domain.Slot{
		{
			ID:     "00000000-0000-0000-0000-000000000702",
			RoomID: room.ID,
			Start:  start1,
			End:    start1.Add(30 * time.Minute),
		},
		{
			ID:     "00000000-0000-0000-0000-000000000703",
			RoomID: room.ID,
			Start:  start2,
			End:    start2.Add(30 * time.Minute),
		},
	}
	if err := slotRepo.BulkUpsert(context.Background(), slots); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	userCreatedAt := time.Now().UTC()
	users := []domain.User{
		{
			ID:        "00000000-0000-0000-0000-000000000704",
			Email:     "listall1@example.com",
			Role:      domain.RoleUser,
			CreatedAt: &userCreatedAt,
		},
		{
			ID:        "00000000-0000-0000-0000-000000000705",
			Email:     "listall2@example.com",
			Role:      domain.RoleUser,
			CreatedAt: &userCreatedAt,
		},
	}
	for _, user := range users {
		if err := userRepo.Upsert(context.Background(), user); err != nil {
			t.Fatalf("Upsert user error = %v", err)
		}
	}

	bookings := []domain.Booking{
		{
			ID:     "00000000-0000-0000-0000-000000000706",
			SlotID: slots[0].ID,
			UserID: users[0].ID,
			Status: domain.BookingStatusActive,
		},
		{
			ID:     "00000000-0000-0000-0000-000000000707",
			SlotID: slots[1].ID,
			UserID: users[1].ID,
			Status: domain.BookingStatusCancelled,
		},
	}
	for _, booking := range bookings {
		if _, err := bookingRepo.Create(context.Background(), booking); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	items, total, err := bookingRepo.ListAll(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}
