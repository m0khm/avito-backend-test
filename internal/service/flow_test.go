package service

import (
	"context"
	"testing"
	"time"

	"room-booking-service/internal/domain"
)

func TestServiceFlow_CreateRoomScheduleBookingFlow(t *testing.T) {
	ctx := context.Background()
	rooms := newFakeRoomRepo()
	schedules := newFakeScheduleRepo()
	slots := newFakeSlotRepo()
	bookings := newFakeBookingRepo()

	roomService := NewRoomService(rooms)
	slotService := NewSlotService(rooms, schedules, slots, 30)
	scheduleService := NewScheduleService(rooms, schedules, slotService, 30)
	bookingService := NewBookingService(bookings, slots, fakeTxManager{}, fakeConference{})

	room, err := roomService.Create(ctx, "Focus room", nil, nil)
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	tomorrow := time.Now().UTC().AddDate(0, 0, 1)
	isoDay := int(tomorrow.Weekday())
	if isoDay == 0 {
		isoDay = 7
	}
	_, err = scheduleService.Create(ctx, room.ID, domainSchedule(isoDay))
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	available, err := slotService.ListAvailableByRoomAndDate(ctx, room.ID, tomorrow.Format("2006-01-02"))
	if err != nil {
		t.Fatalf("list available slots: %v", err)
	}
	if len(available) == 0 {
		t.Fatal("expected generated slots")
	}

	booking, err := bookingService.Create(ctx, available[0].ID, "user-1", false)
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}
	if booking.SlotID != available[0].ID {
		t.Fatalf("expected booked slot %s, got %s", available[0].ID, booking.SlotID)
	}
}

func TestServiceFlow_CancelBookingFlow(t *testing.T) {
	ctx := context.Background()
	rooms := newFakeRoomRepo()
	schedules := newFakeScheduleRepo()
	slots := newFakeSlotRepo()
	bookings := newFakeBookingRepo()

	roomService := NewRoomService(rooms)
	slotService := NewSlotService(rooms, schedules, slots, 30)
	scheduleService := NewScheduleService(rooms, schedules, slotService, 30)
	bookingService := NewBookingService(bookings, slots, fakeTxManager{}, fakeConference{})

	room, _ := roomService.Create(ctx, "Quiet room", nil, nil)
	tomorrow := time.Now().UTC().AddDate(0, 0, 1)
	isoDay := int(tomorrow.Weekday())
	if isoDay == 0 {
		isoDay = 7
	}
	_, _ = scheduleService.Create(ctx, room.ID, domainSchedule(isoDay))
	available, _ := slotService.ListAvailableByRoomAndDate(ctx, room.ID, tomorrow.Format("2006-01-02"))
	booking, _ := bookingService.Create(ctx, available[0].ID, "user-1", false)

	cancelled, err := bookingService.Cancel(ctx, booking.ID, "user-1")
	if err != nil {
		t.Fatalf("cancel booking: %v", err)
	}
	if cancelled.Status != "cancelled" {
		t.Fatalf("expected cancelled status, got %s", cancelled.Status)
	}
}

func domainSchedule(day int) domain.Schedule {
	return domain.Schedule{DaysOfWeek: []int{day}, StartTime: "09:00", EndTime: "10:00"}
}
