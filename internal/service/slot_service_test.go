package service

import (
	"context"
	"testing"
	"time"

	"room-booking-service/internal/domain"
)

func TestGenerateForSchedule(t *testing.T) {
	rooms := newFakeRoomRepo()
	schedules := newFakeScheduleRepo()
	slots := newFakeSlotRepo()
	svc := NewSlotService(rooms, schedules, slots, 7)

	startDate := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC) // Monday
	err := svc.GenerateForSchedule(context.Background(), domain.Schedule{
		RoomID:     "room-1",
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "10:30",
	}, startDate, 1)
	if err != nil {
		t.Fatalf("GenerateForSchedule() error = %v", err)
	}
	if len(slots.slots) != 3 {
		t.Fatalf("expected 3 generated slots, got %d", len(slots.slots))
	}
}
