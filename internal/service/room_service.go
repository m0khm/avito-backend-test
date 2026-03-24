package service

import (
	"context"
	"fmt"
	"strings"

	"room-booking-service/internal/domain"
	apperrors "room-booking-service/internal/errors"

	"github.com/google/uuid"
)

type roomRepo interface {
	Create(ctx context.Context, room domain.Room) (*domain.Room, error)
	List(ctx context.Context) ([]domain.Room, error)
	Exists(ctx context.Context, roomID string) (bool, error)
}

type RoomService struct {
	rooms roomRepo
}

func NewRoomService(rooms roomRepo) *RoomService {
	return &RoomService{rooms: rooms}
}

func (s *RoomService) Create(ctx context.Context, name string, description *string, capacity *int) (*domain.Room, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, apperrors.New(apperrors.ErrInvalidRequest.Code, "name is required", 400)
	}
	if capacity != nil && *capacity <= 0 {
		return nil, apperrors.New(apperrors.ErrInvalidRequest.Code, "capacity must be positive", 400)
	}
	room := domain.Room{ID: uuid.NewString(), Name: name, Description: description, Capacity: capacity}
	created, err := s.rooms.Create(ctx, room)
	if err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	return created, nil
}

func (s *RoomService) List(ctx context.Context) ([]domain.Room, error) {
	return s.rooms.List(ctx)
}
