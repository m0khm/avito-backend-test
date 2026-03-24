package service

import (
	"context"
	"testing"

	"room-booking-service/internal/domain"
)

type roomListRepo struct {
	list []domain.Room
}

func (r *roomListRepo) Create(ctx context.Context, room domain.Room) (*domain.Room, error) {
	copy := room
	return &copy, nil
}

func (r *roomListRepo) List(ctx context.Context) ([]domain.Room, error) {
	return r.list, nil
}

func (r *roomListRepo) Exists(ctx context.Context, roomID string) (bool, error) {
	return true, nil
}

func TestRoomService_List(t *testing.T) {
	repo := &roomListRepo{
		list: []domain.Room{
			{ID: "r1", Name: "Alpha"},
			{ID: "r2", Name: "Beta"},
		},
	}
	svc := NewRoomService(repo)

	rooms, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(rooms) != 2 {
		t.Fatalf("expected 2 rooms, got %d", len(rooms))
	}
}