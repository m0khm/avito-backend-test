package repo

import (
	"context"
	"testing"
	"time"

	"room-booking-service/internal/domain"
)

func TestSlotRepository_BulkUpsertAndGetByID(t *testing.T) {
	db := testDB(t)
	roomRepo := NewRoomRepository(db)
	slotRepo := NewSlotRepository(db)

	room := domain.Room{ID: "00000000-0000-0000-0000-000000000401", Name: "Slot Room"}
	if _, err := roomRepo.Create(context.Background(), room); err != nil {
		t.Fatalf("Create room error = %v", err)
	}

	start := time.Now().UTC().Add(2 * time.Hour).Truncate(30 * time.Minute)
	slot := domain.Slot{ID: "00000000-0000-0000-0000-000000000402", RoomID: room.ID, Start: start, End: start.Add(30 * time.Minute)}
	if err := slotRepo.BulkUpsert(context.Background(), []domain.Slot{slot}); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	got, err := slotRepo.GetByID(context.Background(), slot.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != slot.ID {
		t.Fatalf("expected slot id %s, got %s", slot.ID, got.ID)
	}
}

func TestSlotRepository_BulkUpsert_IsIdempotent(t *testing.T) {
	db := testDB(t)
	roomRepo := NewRoomRepository(db)
	slotRepo := NewSlotRepository(db)

	room := domain.Room{ID: "00000000-0000-0000-0000-000000000411", Name: "Idempotent Room"}
	if _, err := roomRepo.Create(context.Background(), room); err != nil {
		t.Fatalf("Create room error = %v", err)
	}

	start := time.Now().UTC().Add(3 * time.Hour).Truncate(30 * time.Minute)
	slot := domain.Slot{ID: "00000000-0000-0000-0000-000000000412", RoomID: room.ID, Start: start, End: start.Add(30 * time.Minute)}
	if err := slotRepo.BulkUpsert(context.Background(), []domain.Slot{slot}); err != nil {
		t.Fatalf("first BulkUpsert() error = %v", err)
	}
	if err := slotRepo.BulkUpsert(context.Background(), []domain.Slot{slot}); err != nil {
		t.Fatalf("second BulkUpsert() error = %v", err)
	}

	items, err := slotRepo.ListAvailableByRoomAndDate(context.Background(), room.ID, start.Truncate(24*time.Hour))
	if err != nil {
		t.Fatalf("ListAvailableByRoomAndDate() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 slot after duplicate upsert, got %d", len(items))
	}
}

func TestSlotRepository_ListAvailableByRoomAndDate_ExcludesBookedSlots(t *testing.T) {
	db := testDB(t)

	roomRepo := NewRoomRepository(db)
	slotRepo := NewSlotRepository(db)
	userRepo := NewUserRepository(db)
	bookingRepo := NewBookingRepository(db)

	room := domain.Room{
		ID:   "00000000-0000-0000-0000-000000000411",
		Name: "Repo Slot Room",
	}
	if _, err := roomRepo.Create(context.Background(), room); err != nil {
		t.Fatalf("Create room error = %v", err)
	}

	// Делаем слоты заведомо в будущем, чтобы не упереться в фильтр s.start_at >= NOW().
	now := time.Now().UTC()
	targetDay := now.Add(24 * time.Hour)
	dayStart := time.Date(targetDay.Year(), targetDay.Month(), targetDay.Day(), 0, 0, 0, 0, time.UTC)

	slot1Start := dayStart.Add(10 * time.Hour)
	slot2Start := dayStart.Add(10*time.Hour + 30*time.Minute)

	slots := []domain.Slot{
		{
			ID:     "00000000-0000-0000-0000-000000000412",
			RoomID: room.ID,
			Start:  slot1Start,
			End:    slot1Start.Add(30 * time.Minute),
		},
		{
			ID:     "00000000-0000-0000-0000-000000000413",
			RoomID: room.ID,
			Start:  slot2Start,
			End:    slot2Start.Add(30 * time.Minute),
		},
	}
	if err := slotRepo.BulkUpsert(context.Background(), slots); err != nil {
		t.Fatalf("BulkUpsert() error = %v", err)
	}

	userCreatedAt := time.Now().UTC()
	user := domain.User{
		ID:        "00000000-0000-0000-0000-000000000414",
		Email:     "slot-test-user@example.com",
		Role:      domain.RoleUser,
		CreatedAt: &userCreatedAt,
	}
	if err := userRepo.Upsert(context.Background(), user); err != nil {
		t.Fatalf("Upsert user error = %v", err)
	}

	booking := domain.Booking{
		ID:     "00000000-0000-0000-0000-000000000415",
		SlotID: slots[0].ID,
		UserID: user.ID,
		Status: domain.BookingStatusActive,
	}
	if _, err := bookingRepo.Create(context.Background(), booking); err != nil {
		t.Fatalf("Create booking error = %v", err)
	}

	got, err := slotRepo.ListAvailableByRoomAndDate(context.Background(), room.ID, dayStart)
	if err != nil {
		t.Fatalf("ListAvailableByRoomAndDate() error = %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected only unbooked slot, got %+v", got)
	}
	if got[0].ID != slots[1].ID {
		t.Fatalf("expected slot %s, got %s", slots[1].ID, got[0].ID)
	}
}
