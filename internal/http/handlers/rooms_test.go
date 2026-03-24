package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"room-booking-service/internal/domain"
	"room-booking-service/internal/service"
)

type roomsTestRepo struct {
	created *domain.Room
	list    []domain.Room
}

func (r *roomsTestRepo) Create(ctx context.Context, room domain.Room) (*domain.Room, error) {
	r.created = &room
	copy := room
	return &copy, nil
}

func (r *roomsTestRepo) List(ctx context.Context) ([]domain.Room, error) {
	return r.list, nil
}

func (r *roomsTestRepo) Exists(ctx context.Context, roomID string) (bool, error) {
	return true, nil
}

func TestListRooms(t *testing.T) {
	repo := &roomsTestRepo{
		list: []domain.Room{
			{ID: "room-1", Name: "Alpha"},
			{ID: "room-2", Name: "Beta"},
		},
	}
	h := &Handler{rooms: service.NewRoomService(repo)}

	req := httptest.NewRequest(http.MethodGet, "/rooms/list", nil)
	rec := httptest.NewRecorder()

	h.ListRooms(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"Alpha"`) {
		t.Fatalf("expected response to contain room name, body=%s", rec.Body.String())
	}
}

func TestCreateRoom(t *testing.T) {
	repo := &roomsTestRepo{}
	h := &Handler{rooms: service.NewRoomService(repo)}

	body := []byte(`{"name":"Alpha","description":"Test room","capacity":6}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateRoom(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if repo.created == nil {
		t.Fatal("expected room to be created")
	}
	if repo.created.Name != "Alpha" {
		t.Fatalf("expected created name Alpha, got %s", repo.created.Name)
	}
}

func TestCreateRoom_InvalidBody(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/rooms/create", bytes.NewReader([]byte(`{bad json}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateRoom(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}
}