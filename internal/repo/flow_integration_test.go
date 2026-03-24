package repo

import (
	"context"
	"testing"
	"time"

	"room-booking-service/internal/domain"
	"room-booking-service/internal/service"
	"room-booking-service/internal/tx"
)

func isoWeekday(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func TestIntegration_CreateRoomScheduleAndBooking(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)

	roomRepo := NewRoomRepository(db)
	scheduleRepo := NewScheduleRepository(db)
	slotRepo := NewSlotRepository(db)
	bookingRepo := NewBookingRepository(db)
	userRepo := NewUserRepository(db)

	roomService := service.NewRoomService(roomRepo)
	slotService := service.NewSlotService(roomRepo, scheduleRepo, slotRepo, 7)
	scheduleService := service.NewScheduleService(roomRepo, scheduleRepo, slotService, 7)
	bookingService := service.NewBookingService(bookingRepo, slotRepo, tx.NewManager(db), nil)

	user := domain.User{
		ID:    "22222222-2222-2222-2222-222222222222",
		Email: "integration-user@example.com",
		Role:  domain.RoleUser,
	}
	if err := userRepo.Upsert(ctx, user); err != nil {
		t.Fatalf("Upsert user error = %v", err)
	}

	room, err := roomService.Create(ctx, "Integration Room", nil, nil)
	if err != nil {
		t.Fatalf("Create room error = %v", err)
	}

	targetDate := time.Now().UTC().Add(24 * time.Hour)
	_, err = scheduleService.Create(ctx, room.ID, domain.Schedule{
		DaysOfWeek: []int{isoWeekday(targetDate)},
		StartTime:  "09:00",
		EndTime:    "10:00",
	})
	if err != nil {
		t.Fatalf("Create schedule error = %v", err)
	}

	slots, err := slotService.ListAvailableByRoomAndDate(ctx, room.ID, targetDate.Format("2006-01-02"))
	if err != nil {
		t.Fatalf("ListAvailableByRoomAndDate error = %v", err)
	}
	if len(slots) != 2 {
		t.Fatalf("expected 2 generated slots, got %d", len(slots))
	}

	booking, err := bookingService.Create(ctx, slots[0].ID, user.ID, false)
	if err != nil {
		t.Fatalf("Create booking error = %v", err)
	}
	if booking.Status != domain.BookingStatusActive {
		t.Fatalf("expected active booking, got %s", booking.Status)
	}
	if booking.SlotID != slots[0].ID {
		t.Fatalf("expected booking slot %s, got %s", slots[0].ID, booking.SlotID)
	}

	remaining, err := slotService.ListAvailableByRoomAndDate(ctx, room.ID, targetDate.Format("2006-01-02"))
	if err != nil {
		t.Fatalf("ListAvailableByRoomAndDate after booking error = %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 remaining available slot, got %d", len(remaining))
	}
	if remaining[0].ID != slots[1].ID {
		t.Fatalf("expected remaining slot %s, got %s", slots[1].ID, remaining[0].ID)
	}
}

func TestIntegration_CancelBooking_IsIdempotent(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)

	roomRepo := NewRoomRepository(db)
	scheduleRepo := NewScheduleRepository(db)
	slotRepo := NewSlotRepository(db)
	bookingRepo := NewBookingRepository(db)
	userRepo := NewUserRepository(db)

	roomService := service.NewRoomService(roomRepo)
	slotService := service.NewSlotService(roomRepo, scheduleRepo, slotRepo, 7)
	scheduleService := service.NewScheduleService(roomRepo, scheduleRepo, slotService, 7)
	bookingService := service.NewBookingService(bookingRepo, slotRepo, tx.NewManager(db), nil)

	user := domain.User{
		ID:    "33333333-3333-3333-3333-333333333333",
		Email: "cancel-user@example.com",
		Role:  domain.RoleUser,
	}
	if err := userRepo.Upsert(ctx, user); err != nil {
		t.Fatalf("Upsert user error = %v", err)
	}

	room, err := roomService.Create(ctx, "Cancel Flow Room", nil, nil)
	if err != nil {
		t.Fatalf("Create room error = %v", err)
	}

	targetDate := time.Now().UTC().Add(24 * time.Hour)
	_, err = scheduleService.Create(ctx, room.ID, domain.Schedule{
		DaysOfWeek: []int{isoWeekday(targetDate)},
		StartTime:  "09:00",
		EndTime:    "10:00",
	})
	if err != nil {
		t.Fatalf("Create schedule error = %v", err)
	}

	slots, err := slotService.ListAvailableByRoomAndDate(ctx, room.ID, targetDate.Format("2006-01-02"))
	if err != nil {
		t.Fatalf("ListAvailableByRoomAndDate error = %v", err)
	}
	if len(slots) == 0 {
		t.Fatal("expected at least one generated slot")
	}

	booking, err := bookingService.Create(ctx, slots[0].ID, user.ID, false)
	if err != nil {
		t.Fatalf("Create booking error = %v", err)
	}

	cancelled, err := bookingService.Cancel(ctx, booking.ID, user.ID)
	if err != nil {
		t.Fatalf("first Cancel error = %v", err)
	}
	if cancelled.Status != domain.BookingStatusCancelled {
		t.Fatalf("expected cancelled status, got %s", cancelled.Status)
	}

	cancelledAgain, err := bookingService.Cancel(ctx, booking.ID, user.ID)
	if err != nil {
		t.Fatalf("second Cancel error = %v", err)
	}
	if cancelledAgain.Status != domain.BookingStatusCancelled {
		t.Fatalf("expected cancelled status on second cancel, got %s", cancelledAgain.Status)
	}

	availableAgain, err := slotService.ListAvailableByRoomAndDate(ctx, room.ID, targetDate.Format("2006-01-02"))
	if err != nil {
		t.Fatalf("ListAvailableByRoomAndDate after cancel error = %v", err)
	}
	if len(availableAgain) != 2 {
		t.Fatalf("expected 2 available slots after cancel, got %d", len(availableAgain))
	}
}