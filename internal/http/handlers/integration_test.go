package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"room-booking-service/internal/domain"
	authmw "room-booking-service/internal/http/middleware"
	"room-booking-service/internal/service"
)

type integrationUserRepo struct {
	mu      sync.Mutex
	byEmail map[string]domain.User
}

func (r *integrationUserRepo) Upsert(_ context.Context, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.byEmail == nil {
		r.byEmail = map[string]domain.User{}
	}
	r.byEmail[user.Email] = user
	return nil
}
func (r *integrationUserRepo) Create(_ context.Context, user domain.User) (*domain.User, error) {
	if err := r.Upsert(context.Background(), user); err != nil {
		return nil, err
	}
	copy := user
	return &copy, nil
}
func (r *integrationUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.byEmail[email]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	copy := user
	return &copy, nil
}

type integrationRoomRepo struct {
	mu    sync.Mutex
	rooms map[string]domain.Room
}

func (r *integrationRoomRepo) Create(_ context.Context, room domain.Room) (*domain.Room, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.rooms == nil {
		r.rooms = map[string]domain.Room{}
	}
	r.rooms[room.ID] = room
	copy := room
	return &copy, nil
}
func (r *integrationRoomRepo) List(_ context.Context) ([]domain.Room, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]domain.Room, 0, len(r.rooms))
	for _, room := range r.rooms {
		items = append(items, room)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, nil
}
func (r *integrationRoomRepo) Exists(_ context.Context, roomID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.rooms[roomID]
	return ok, nil
}

type integrationScheduleRepo struct {
	mu        sync.Mutex
	schedules map[string]domain.Schedule
	byRoomID  map[string]string
}

func (r *integrationScheduleRepo) Create(_ context.Context, schedule domain.Schedule) (*domain.Schedule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.schedules == nil {
		r.schedules = map[string]domain.Schedule{}
		r.byRoomID = map[string]string{}
	}
	r.schedules[schedule.ID] = schedule
	r.byRoomID[schedule.RoomID] = schedule.ID
	copy := schedule
	return &copy, nil
}
func (r *integrationScheduleRepo) GetByRoomID(_ context.Context, roomID string) (*domain.Schedule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byRoomID[roomID]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	s := r.schedules[id]
	copy := s
	return &copy, nil
}
func (r *integrationScheduleRepo) List(_ context.Context) ([]domain.Schedule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]domain.Schedule, 0, len(r.schedules))
	for _, sched := range r.schedules {
		items = append(items, sched)
	}
	return items, nil
}

type integrationSlotRepo struct {
	mu       sync.Mutex
	slots    map[string]domain.Slot
	bookings *integrationBookingRepo
}

func (r *integrationSlotRepo) BulkUpsert(_ context.Context, slots []domain.Slot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.slots == nil {
		r.slots = map[string]domain.Slot{}
	}
	for _, slot := range slots {
		if _, exists := r.slots[slot.ID]; !exists {
			r.slots[slot.ID] = slot
		}
	}
	return nil
}
func (r *integrationSlotRepo) ListAvailableByRoomAndDate(_ context.Context, roomID string, date time.Time) ([]domain.Slot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	start, end := date, date.Add(24*time.Hour)
	items := make([]domain.Slot, 0)
	for _, slot := range r.slots {
		if slot.RoomID != roomID || slot.Start.Before(start) || !slot.Start.Before(end) || slot.Start.Before(time.Now().UTC()) {
			continue
		}
		if r.bookings != nil && r.bookings.hasActiveBooking(slot.ID) {
			continue
		}
		items = append(items, slot)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Start.Before(items[j].Start) })
	return items, nil
}
func (r *integrationSlotRepo) GetByID(_ context.Context, slotID string) (*domain.Slot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	slot, ok := r.slots[slotID]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	copy := slot
	return &copy, nil
}

type integrationBookingRepo struct {
	mu       sync.Mutex
	bookings map[string]domain.Booking
	slots    *integrationSlotRepo
}

func (r *integrationBookingRepo) hasActiveBooking(slotID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, booking := range r.bookings {
		if booking.SlotID == slotID && booking.Status == domain.BookingStatusActive {
			return true
		}
	}
	return false
}

func (r *integrationBookingRepo) Create(_ context.Context, booking domain.Booking) (*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.bookings == nil {
		r.bookings = map[string]domain.Booking{}
	}
	for _, existing := range r.bookings {
		if existing.SlotID == booking.SlotID && existing.Status == domain.BookingStatusActive {
			return nil, errDuplicate
		}
	}
	r.bookings[booking.ID] = booking
	copy := booking
	return &copy, nil
}
func (r *integrationBookingRepo) UpdateConferenceLink(_ context.Context, bookingID string, conferenceLink *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	booking := r.bookings[bookingID]
	booking.ConferenceLink = conferenceLink
	r.bookings[bookingID] = booking
	return nil
}
func (r *integrationBookingRepo) ListAll(_ context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]domain.Booking, 0, len(r.bookings))
	for _, booking := range r.bookings {
		items = append(items, booking)
	}
	return items, len(items), nil
}
func (r *integrationBookingRepo) ListFutureByUser(_ context.Context, userID string) ([]domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]domain.Booking, 0)
	for _, booking := range r.bookings {
		if booking.UserID == userID && booking.Status == domain.BookingStatusActive {
			items = append(items, booking)
		}
	}
	return items, nil
}
func (r *integrationBookingRepo) GetByID(_ context.Context, bookingID string) (*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	booking, ok := r.bookings[bookingID]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	copy := booking
	return &copy, nil
}
func (r *integrationBookingRepo) Cancel(_ context.Context, bookingID string) (*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	booking := r.bookings[bookingID]
	booking.Status = domain.BookingStatusCancelled
	r.bookings[bookingID] = booking
	copy := booking
	return &copy, nil
}

type integrationTxManager struct{}

func (integrationTxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type integrationConference struct{}

func (integrationConference) CreateConferenceLink(_ context.Context, _, _, _ string) (*string, error) {
	link := "https://meet.mock/test"
	return &link, nil
}

var errDuplicate = serviceErr("duplicate key value violates unique constraint uq_active_booking_slot")

type serviceErr string

func (e serviceErr) Error() string { return string(e) }

func newIntegrationRouter(t *testing.T) http.Handler {
	t.Helper()
	users := &integrationUserRepo{}
	rooms := &integrationRoomRepo{}
	schedules := &integrationScheduleRepo{}
	bookings := &integrationBookingRepo{bookings: map[string]domain.Booking{}}
	slots := &integrationSlotRepo{slots: map[string]domain.Slot{}, bookings: bookings}
	bookings.slots = slots

	authSvc := service.NewAuthService(users, "secret", "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	roomSvc := service.NewRoomService(rooms)
	slotSvc := service.NewSlotService(rooms, schedules, slots, 7)
	scheduleSvc := service.NewScheduleService(rooms, schedules, slotSvc, 7)
	bookingSvc := service.NewBookingService(bookings, slots, integrationTxManager{}, integrationConference{})

	h := New(authSvc, roomSvc, scheduleSvc, slotSvc, bookingSvc)
	r := chi.NewRouter()
	auth := authmw.NewAuthMiddleware("secret")
	r.Post("/dummyLogin", h.DummyLogin)
	r.Group(func(protected chi.Router) {
		protected.Use(auth.RequireAuth)
		protected.Get("/rooms/list", h.ListRooms)
		protected.Get("/rooms/{roomId}/slots/list", h.ListSlots)
		protected.With(authmw.RequireRole("admin")).Post("/rooms/create", h.CreateRoom)
		protected.With(authmw.RequireRole("admin")).Post("/rooms/{roomId}/schedule/create", h.CreateSchedule)
		protected.With(authmw.RequireRole("user")).Post("/bookings/create", h.CreateBooking)
		protected.With(authmw.RequireRole("user")).Post("/bookings/{bookingId}/cancel", h.CancelBooking)
	})
	return r
}

func authToken(t *testing.T, router http.Handler, role string) string {
	t.Helper()
	body := []byte(`{"role":"` + role + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("dummyLogin failed: %d %s", rec.Code, rec.Body.String())
	}
	var resp tokenResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal token response: %v", err)
	}
	return resp.Token
}

func TestHTTPFlow_CreateAndCancelBooking(t *testing.T) {
	router := newIntegrationRouter(t)
	adminToken := authToken(t, router, "admin")
	userToken := authToken(t, router, "user")

	createRoomReq := httptest.NewRequest(http.MethodPost, "/rooms/create", bytes.NewReader([]byte(`{"name":"Focus room"}`)))
	createRoomReq.Header.Set("Content-Type", "application/json")
	createRoomReq.Header.Set("Authorization", "Bearer "+adminToken)
	createRoomRec := httptest.NewRecorder()
	router.ServeHTTP(createRoomRec, createRoomReq)
	if createRoomRec.Code != http.StatusCreated {
		t.Fatalf("create room failed: %d %s", createRoomRec.Code, createRoomRec.Body.String())
	}
	var roomResp roomResponse
	if err := json.Unmarshal(createRoomRec.Body.Bytes(), &roomResp); err != nil {
		t.Fatalf("unmarshal room response: %v", err)
	}

	tomorrow := time.Now().UTC().AddDate(0, 0, 1)
	isoDay := int(tomorrow.Weekday())
	if isoDay == 0 {
		isoDay = 7
	}
	schedBody, _ := json.Marshal(struct {
	RoomID     string `json:"roomId"`
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
	}{
	RoomID:     roomResp.Room.ID,
	DaysOfWeek: []int{isoDay},
	StartTime:  "09:00",
	EndTime:    "10:00",
	})
	scheduleReq := httptest.NewRequest(http.MethodPost, "/rooms/"+roomResp.Room.ID+"/schedule/create", bytes.NewReader(schedBody))
	scheduleReq.Header.Set("Content-Type", "application/json")
	scheduleReq.Header.Set("Authorization", "Bearer "+adminToken)
	scheduleRec := httptest.NewRecorder()
	router.ServeHTTP(scheduleRec, scheduleReq)
	if scheduleRec.Code != http.StatusCreated {
		t.Fatalf("create schedule failed: %d %s", scheduleRec.Code, scheduleRec.Body.String())
	}

	dateParam := tomorrow.Format("2006-01-02")
	listSlotsReq := httptest.NewRequest(http.MethodGet, "/rooms/"+roomResp.Room.ID+"/slots/list?date="+dateParam, nil)
	listSlotsReq.Header.Set("Authorization", "Bearer "+userToken)
	listSlotsRec := httptest.NewRecorder()
	router.ServeHTTP(listSlotsRec, listSlotsReq)
	if listSlotsRec.Code != http.StatusOK {
		t.Fatalf("list slots failed: %d %s", listSlotsRec.Code, listSlotsRec.Body.String())
	}
	var slotsResp slotsResponse
	if err := json.Unmarshal(listSlotsRec.Body.Bytes(), &slotsResp); err != nil {
		t.Fatalf("unmarshal slots response: %v", err)
	}
	if len(slotsResp.Slots) == 0 {
		t.Fatal("expected generated slots")
	}

	createBookingBody := []byte(`{"slotId":"` + slotsResp.Slots[0].ID + `","createConferenceLink":true}`)
	createBookingReq := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewReader(createBookingBody))
	createBookingReq.Header.Set("Content-Type", "application/json")
	createBookingReq.Header.Set("Authorization", "Bearer "+userToken)
	createBookingRec := httptest.NewRecorder()
	router.ServeHTTP(createBookingRec, createBookingReq)
	if createBookingRec.Code != http.StatusCreated {
		t.Fatalf("create booking failed: %d %s", createBookingRec.Code, createBookingRec.Body.String())
	}
	var bookingResp bookingResponse
	if err := json.Unmarshal(createBookingRec.Body.Bytes(), &bookingResp); err != nil {
		t.Fatalf("unmarshal booking response: %v", err)
	}
	if bookingResp.Booking.ConferenceLink == nil {
		t.Fatal("expected conference link")
	}

	cancelReq := httptest.NewRequest(http.MethodPost, "/bookings/"+bookingResp.Booking.ID+"/cancel", nil)
	cancelReq.Header.Set("Authorization", "Bearer "+userToken)
	cancelRec := httptest.NewRecorder()
	router.ServeHTTP(cancelRec, cancelReq)
	if cancelRec.Code != http.StatusOK {
		t.Fatalf("cancel booking failed: %d %s", cancelRec.Code, cancelRec.Body.String())
	}
	var cancelled bookingResponse
	if err := json.Unmarshal(cancelRec.Body.Bytes(), &cancelled); err != nil {
		t.Fatalf("unmarshal cancelled response: %v", err)
	}
	if cancelled.Booking.Status != domain.BookingStatusCancelled {
		t.Fatalf("expected cancelled booking, got %+v", cancelled.Booking)
	}
}
