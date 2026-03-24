package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"room-booking-service/internal/domain"
	"room-booking-service/internal/service"
)

type scheduleHandlerRoomRepo struct {
	exists bool
}

func (r *scheduleHandlerRoomRepo) Create(ctx context.Context, room domain.Room) (*domain.Room, error) {
	copy := room
	return &copy, nil
}

func (r *scheduleHandlerRoomRepo) List(ctx context.Context) ([]domain.Room, error) {
	return nil, nil
}

func (r *scheduleHandlerRoomRepo) Exists(ctx context.Context, roomID string) (bool, error) {
	return r.exists, nil
}

type scheduleHandlerScheduleRepo struct {
	created       *domain.Schedule
	duplicateOnce bool
}

func (r *scheduleHandlerScheduleRepo) Create(ctx context.Context, schedule domain.Schedule) (*domain.Schedule, error) {
	if r.duplicateOnce {
		return nil,  errors.New("duplicate key")
	}

	r.created = &schedule
	copy := schedule
	now := time.Now().UTC()
	copy.CreatedAt = &now
	return &copy, nil
}

func (r *scheduleHandlerScheduleRepo) GetByRoomID(ctx context.Context, roomID string) (*domain.Schedule, error) {
	return nil, nil
}

func (r *scheduleHandlerScheduleRepo) List(ctx context.Context) ([]domain.Schedule, error) {
	return nil, nil
}

type scheduleHandlerSlotGen struct {
	called bool
}

func (g *scheduleHandlerSlotGen) GenerateForSchedule(ctx context.Context, schedule domain.Schedule, fromDate time.Time, horizonDays int) error {
	g.called = true
	return nil
}

func (g *scheduleHandlerSlotGen) ExtendAll(ctx context.Context) error {
	return nil
}

func withRoomIDParam(r *http.Request, roomID string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("roomId", roomID)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestCreateSchedule_Success(t *testing.T) {
	const roomID = "11111111-1111-1111-1111-111111111111"

	roomRepo := &scheduleHandlerRoomRepo{exists: true}
	scheduleRepo := &scheduleHandlerScheduleRepo{}
	slotGen := &scheduleHandlerSlotGen{}

	h := &Handler{
		schedules: service.NewScheduleService(roomRepo, scheduleRepo, slotGen, 7),
	}

	body := []byte(`{"daysOfWeek":[1,2,3],"startTime":"09:00","endTime":"18:00"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID+"/schedule/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withRoomIDParam(req, roomID)

	rec := httptest.NewRecorder()
	h.CreateSchedule(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if scheduleRepo.created == nil {
		t.Fatal("expected schedule to be created")
	}
	if scheduleRepo.created.RoomID != roomID {
		t.Fatalf("expected roomID %s, got %s", roomID, scheduleRepo.created.RoomID)
	}
	if !slotGen.called {
		t.Fatal("expected slot generator to be called")
	}
	if !strings.Contains(rec.Body.String(), `"daysOfWeek":[1,2,3]`) {
		t.Fatalf("expected response to contain created schedule, body=%s", rec.Body.String())
	}
}

func TestCreateSchedule_InvalidBody(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/rooms/room-id/schedule/create", bytes.NewReader([]byte(`{bad json}`)))
	req.Header.Set("Content-Type", "application/json")
	req = withRoomIDParam(req, "11111111-1111-1111-1111-111111111111")

	rec := httptest.NewRecorder()
	h.CreateSchedule(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateSchedule_RoomNotFound(t *testing.T) {
	const roomID = "11111111-1111-1111-1111-111111111111"

	roomRepo := &scheduleHandlerRoomRepo{exists: false}
	scheduleRepo := &scheduleHandlerScheduleRepo{}
	slotGen := &scheduleHandlerSlotGen{}

	h := &Handler{
		schedules: service.NewScheduleService(roomRepo, scheduleRepo, slotGen, 7),
	}

	body := []byte(`{"daysOfWeek":[1,2,3],"startTime":"09:00","endTime":"18:00"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID+"/schedule/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withRoomIDParam(req, roomID)

	rec := httptest.NewRecorder()
	h.CreateSchedule(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"ROOM_NOT_FOUND"`) {
		t.Fatalf("expected ROOM_NOT_FOUND error, body=%s", rec.Body.String())
	}
}

func TestCreateSchedule_Conflict(t *testing.T) {
	const roomID = "11111111-1111-1111-1111-111111111111"

	roomRepo := &scheduleHandlerRoomRepo{exists: true}
	scheduleRepo := &scheduleHandlerScheduleRepo{duplicateOnce: true}
	slotGen := &scheduleHandlerSlotGen{}

	h := &Handler{
		schedules: service.NewScheduleService(roomRepo, scheduleRepo, slotGen, 7),
	}

	body := []byte(`{"daysOfWeek":[1,2,3],"startTime":"09:00","endTime":"18:00"}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID+"/schedule/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withRoomIDParam(req, roomID)

	rec := httptest.NewRecorder()
	h.CreateSchedule(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"SCHEDULE_EXISTS"`) {
		t.Fatalf("expected SCHEDULE_EXISTS error, body=%s", rec.Body.String())
	}
}