package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"room-booking-service/internal/domain"
	"room-booking-service/internal/service"
)

type slotsHandlerRoomRepo struct {
	exists bool
}

func (r *slotsHandlerRoomRepo) Create(ctx context.Context, room domain.Room) (*domain.Room, error) {
	copy := room
	return &copy, nil
}

func (r *slotsHandlerRoomRepo) List(ctx context.Context) ([]domain.Room, error) {
	return nil, nil
}

func (r *slotsHandlerRoomRepo) Exists(ctx context.Context, roomID string) (bool, error) {
	return r.exists, nil
}

type slotsHandlerScheduleRepo struct {
	schedule *domain.Schedule
	err      error
}

func (r *slotsHandlerScheduleRepo) Create(ctx context.Context, schedule domain.Schedule) (*domain.Schedule, error) {
	copy := schedule
	return &copy, nil
}

func (r *slotsHandlerScheduleRepo) GetByRoomID(ctx context.Context, roomID string) (*domain.Schedule, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.schedule == nil {
		return nil, pgx.ErrNoRows
	}
	return r.schedule, nil
}

func (r *slotsHandlerScheduleRepo) List(ctx context.Context) ([]domain.Schedule, error) {
	return nil, nil
}

type slotsHandlerSlotRepo struct {
	list []domain.Slot
}

func (r *slotsHandlerSlotRepo) BulkUpsert(ctx context.Context, slots []domain.Slot) error {
	return nil
}

func (r *slotsHandlerSlotRepo) ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]domain.Slot, error) {
	return r.list, nil
}

func (r *slotsHandlerSlotRepo) GetByID(ctx context.Context, slotID string) (*domain.Slot, error) {
	return nil, nil
}

func withRoomID(r *http.Request, roomID string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("roomId", roomID)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestListSlots_Success(t *testing.T) {
	const roomID = "11111111-1111-1111-1111-111111111111"

	roomRepo := &slotsHandlerRoomRepo{exists: true}
	scheduleRepo := &slotsHandlerScheduleRepo{
		schedule: &domain.Schedule{
			ID:         "schedule-1",
			RoomID:     roomID,
			DaysOfWeek: []int{2},
			StartTime:  "09:00",
			EndTime:    "18:00",
		},
	}
	slotRepo := &slotsHandlerSlotRepo{
		list: []domain.Slot{
			{
				ID:     "slot-1",
				RoomID: roomID,
				Start:  time.Date(2026, 3, 24, 9, 0, 0, 0, time.UTC),
				End:    time.Date(2026, 3, 24, 9, 30, 0, 0, time.UTC),
			},
		},
	}

	h := &Handler{
		slots: service.NewSlotService(roomRepo, scheduleRepo, slotRepo, 7),
	}

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID+"/slots/list?date=2026-03-24", nil)
	req = withRoomID(req, roomID)

	rec := httptest.NewRecorder()
	h.ListSlots(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"slot-1"`) {
		t.Fatalf("expected response to contain slot, body=%s", rec.Body.String())
	}
}

func TestListSlots_MissingDate(t *testing.T) {
	const roomID = "11111111-1111-1111-1111-111111111111"

	roomRepo := &slotsHandlerRoomRepo{exists: true}
	scheduleRepo := &slotsHandlerScheduleRepo{}
	slotRepo := &slotsHandlerSlotRepo{}

	h := &Handler{
		slots: service.NewSlotService(roomRepo, scheduleRepo, slotRepo, 7),
	}

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID+"/slots/list", nil)
	req = withRoomID(req, roomID)

	rec := httptest.NewRecorder()
	h.ListSlots(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"INVALID_REQUEST"`) {
		t.Fatalf("expected INVALID_REQUEST error, body=%s", rec.Body.String())
	}
}

func TestListSlots_RoomNotFound(t *testing.T) {
	const roomID = "11111111-1111-1111-1111-111111111111"

	roomRepo := &slotsHandlerRoomRepo{exists: false}
	scheduleRepo := &slotsHandlerScheduleRepo{}
	slotRepo := &slotsHandlerSlotRepo{}

	h := &Handler{
		slots: service.NewSlotService(roomRepo, scheduleRepo, slotRepo, 7),
	}

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID+"/slots/list?date=2026-03-24", nil)
	req = withRoomID(req, roomID)

	rec := httptest.NewRecorder()
	h.ListSlots(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"ROOM_NOT_FOUND"`) {
		t.Fatalf("expected ROOM_NOT_FOUND error, body=%s", rec.Body.String())
	}
}

func TestListSlots_NoScheduleReturnsEmptyList(t *testing.T) {
	const roomID = "11111111-1111-1111-1111-111111111111"

	roomRepo := &slotsHandlerRoomRepo{exists: true}
	scheduleRepo := &slotsHandlerScheduleRepo{err: pgx.ErrNoRows}
	slotRepo := &slotsHandlerSlotRepo{}

	h := &Handler{
		slots: service.NewSlotService(roomRepo, scheduleRepo, slotRepo, 7),
	}

	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID+"/slots/list?date=2026-03-24", nil)
	req = withRoomID(req, roomID)

	rec := httptest.NewRecorder()
	h.ListSlots(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"slots":[]`) {
		t.Fatalf("expected empty slots list, body=%s", rec.Body.String())
	}
}