package repo

import (
	"context"
	"testing"

	"room-booking-service/internal/domain"
)

func TestRoomRepository_CreateAndList(t *testing.T) {
	db := testDB(t)
	r := NewRoomRepository(db)

	room := domain.Room{
		ID:   "00000000-0000-0000-0000-000000000101",
		Name: "Repo Test Room",
	}

	created, err := r.Create(context.Background(), room)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected created room id")
	}

	rooms, err := r.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(rooms) == 0 {
		t.Fatal("expected at least one room")
	}
}

func TestRoomRepository_Exists(t *testing.T) {
	db := testDB(t)
	r := NewRoomRepository(db)

	room := domain.Room{
		ID:   "00000000-0000-0000-0000-000000000102",
		Name: "Exists Room",
	}

	_, err := r.Create(context.Background(), room)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	ok, err := r.Exists(context.Background(), room.ID)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !ok {
		t.Fatal("expected room to exist")
	}
}