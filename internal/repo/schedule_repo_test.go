package repo

import (
	"context"
	"testing"

	"room-booking-service/internal/domain"
)

func TestScheduleRepository_CreateAndGetByRoomID(t *testing.T) {
	db := testDB(t)

	roomRepo := NewRoomRepository(db)
	scheduleRepo := NewScheduleRepository(db)

	room := domain.Room{
		ID:   "00000000-0000-0000-0000-000000000301",
		Name: "Schedule Room",
	}
	_, err := roomRepo.Create(context.Background(), room)
	if err != nil {
		t.Fatalf("Create room error = %v", err)
	}

	s := domain.Schedule{
		ID:         "00000000-0000-0000-0000-000000000302",
		RoomID:     room.ID,
		DaysOfWeek: []int{1, 2, 3},
		StartTime:  "09:00",
		EndTime:    "12:00",
	}

	_, err = scheduleRepo.Create(context.Background(), s)
	if err != nil {
		t.Fatalf("Create schedule error = %v", err)
	}

	got, err := scheduleRepo.GetByRoomID(context.Background(), room.ID)
	if err != nil {
		t.Fatalf("GetByRoomID() error = %v", err)
	}
	if got.RoomID != room.ID {
		t.Fatalf("expected room id %s, got %s", room.ID, got.RoomID)
	}
}

func TestScheduleRepository_List(t *testing.T) {
	db := testDB(t)

	roomRepo := NewRoomRepository(db)
	scheduleRepo := NewScheduleRepository(db)

	rooms := []domain.Room{
		{ID: "00000000-0000-0000-0000-000000000401", Name: "Room A"},
		{ID: "00000000-0000-0000-0000-000000000402", Name: "Room B"},
	}
	for _, room := range rooms {
		if _, err := roomRepo.Create(context.Background(), room); err != nil {
			t.Fatalf("Create room error = %v", err)
		}
	}

	schedules := []domain.Schedule{
		{
			ID:         "00000000-0000-0000-0000-000000000403",
			RoomID:     rooms[0].ID,
			DaysOfWeek: []int{1, 3},
			StartTime:  "09:00",
			EndTime:    "11:00",
		},
		{
			ID:         "00000000-0000-0000-0000-000000000404",
			RoomID:     rooms[1].ID,
			DaysOfWeek: []int{2, 4},
			StartTime:  "12:00",
			EndTime:    "14:00",
		},
	}

	for _, s := range schedules {
		if _, err := scheduleRepo.Create(context.Background(), s); err != nil {
			t.Fatalf("Create schedule error = %v", err)
		}
	}

	items, err := scheduleRepo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 schedules, got %d", len(items))
	}

	gotByRoom := map[string]domain.Schedule{}
	for _, item := range items {
		gotByRoom[item.RoomID] = item
	}

	if _, ok := gotByRoom[rooms[0].ID]; !ok {
		t.Fatal("expected schedule for first room")
	}
	if _, ok := gotByRoom[rooms[1].ID]; !ok {
		t.Fatal("expected schedule for second room")
	}
}