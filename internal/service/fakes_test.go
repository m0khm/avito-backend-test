package service

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"room-booking-service/internal/domain"

	"github.com/jackc/pgx/v5"
)

type fakeUserRepo struct {
	mu      sync.Mutex
	byEmail map[string]domain.User
	byID    map[string]domain.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byEmail: map[string]domain.User{}, byID: map[string]domain.User{}}
}

func (r *fakeUserRepo) Upsert(_ context.Context, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byEmail[user.Email] = user
	r.byID[user.ID] = user
	return nil
}

func (r *fakeUserRepo) Create(_ context.Context, user domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byEmail[user.Email]; exists {
		return nil, errors.New("duplicate key")
	}
	now := time.Now().UTC()
	user.CreatedAt = &now
	r.byEmail[user.Email] = user
	r.byID[user.ID] = user
	return &user, nil
}

func (r *fakeUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, exists := r.byEmail[email]
	if !exists {
		return nil, pgx.ErrNoRows
	}
	return &user, nil
}

type fakeRoomRepo struct {
	mu    sync.Mutex
	rooms map[string]domain.Room
}

func newFakeRoomRepo() *fakeRoomRepo { return &fakeRoomRepo{rooms: map[string]domain.Room{}} }

func (r *fakeRoomRepo) Create(_ context.Context, room domain.Room) (*domain.Room, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	room.CreatedAt = &now
	r.rooms[room.ID] = room
	return &room, nil
}
func (r *fakeRoomRepo) List(_ context.Context) ([]domain.Room, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rooms := make([]domain.Room, 0, len(r.rooms))
	for _, room := range r.rooms {
		rooms = append(rooms, room)
	}
	sort.Slice(rooms, func(i, j int) bool { return rooms[i].Name < rooms[j].Name })
	return rooms, nil
}
func (r *fakeRoomRepo) Exists(_ context.Context, roomID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.rooms[roomID]
	return exists, nil
}

type fakeScheduleRepo struct {
	mu        sync.Mutex
	schedules map[string]domain.Schedule
	byRoomID  map[string]string
}

func newFakeScheduleRepo() *fakeScheduleRepo {
	return &fakeScheduleRepo{schedules: map[string]domain.Schedule{}, byRoomID: map[string]string{}}
}

func (r *fakeScheduleRepo) Create(_ context.Context, schedule domain.Schedule) (*domain.Schedule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byRoomID[schedule.RoomID]; exists {
		return nil, errors.New("duplicate key")
	}
	now := time.Now().UTC()
	schedule.CreatedAt = &now
	r.schedules[schedule.ID] = schedule
	r.byRoomID[schedule.RoomID] = schedule.ID
	return &schedule, nil
}
func (r *fakeScheduleRepo) GetByRoomID(_ context.Context, roomID string) (*domain.Schedule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byRoomID[roomID]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	s := r.schedules[id]
	return &s, nil
}
func (r *fakeScheduleRepo) List(_ context.Context) ([]domain.Schedule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make([]domain.Schedule, 0, len(r.schedules))
	for _, s := range r.schedules {
		res = append(res, s)
	}
	return res, nil
}

type fakeSlotRepo struct {
	mu    sync.Mutex
	slots map[string]domain.Slot
}

func newFakeSlotRepo() *fakeSlotRepo { return &fakeSlotRepo{slots: map[string]domain.Slot{}} }

func (r *fakeSlotRepo) BulkUpsert(_ context.Context, slots []domain.Slot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, slot := range slots {
		if _, exists := r.slots[slot.ID]; !exists {
			r.slots[slot.ID] = slot
		}
	}
	return nil
}
func (r *fakeSlotRepo) ListAvailableByRoomAndDate(_ context.Context, roomID string, date time.Time) ([]domain.Slot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	start := date
	end := date.Add(24 * time.Hour)
	res := make([]domain.Slot, 0)
	for _, slot := range r.slots {
		if slot.RoomID == roomID && !slot.Start.Before(start) && slot.Start.Before(end) && !slot.Start.Before(time.Now().UTC()) {
			res = append(res, slot)
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Start.Before(res[j].Start) })
	return res, nil
}
func (r *fakeSlotRepo) GetByID(_ context.Context, slotID string) (*domain.Slot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.slots[slotID]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return &s, nil
}

type fakeBookingRepo struct {
	mu       sync.Mutex
	bookings map[string]domain.Booking
}

func newFakeBookingRepo() *fakeBookingRepo {
	return &fakeBookingRepo{bookings: map[string]domain.Booking{}}
}

func (r *fakeBookingRepo) Create(_ context.Context, booking domain.Booking) (*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.bookings {
		if existing.SlotID == booking.SlotID && existing.Status == domain.BookingStatusActive {
			return nil, errors.New("uq_active_booking_slot")
		}
	}
	now := time.Now().UTC()
	booking.CreatedAt = &now
	r.bookings[booking.ID] = booking
	return &booking, nil
}
func (r *fakeBookingRepo) UpdateConferenceLink(_ context.Context, bookingID string, conferenceLink *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b := r.bookings[bookingID]
	b.ConferenceLink = conferenceLink
	r.bookings[bookingID] = b
	return nil
}

func (r *fakeBookingRepo) ListAll(_ context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]domain.Booking, 0, len(r.bookings))
	for _, b := range r.bookings {
		items = append(items, b)
	}
	return items, len(items), nil
}
func (r *fakeBookingRepo) ListFutureByUser(_ context.Context, userID string) ([]domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make([]domain.Booking, 0)
	for _, b := range r.bookings {
		if b.UserID == userID {
			res = append(res, b)
		}
	}
	return res, nil
}
func (r *fakeBookingRepo) GetByID(_ context.Context, bookingID string) (*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bookings[bookingID]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return &b, nil
}
func (r *fakeBookingRepo) Cancel(_ context.Context, bookingID string) (*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bookings[bookingID]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	b.Status = domain.BookingStatusCancelled
	r.bookings[bookingID] = b
	return &b, nil
}

type fakeTxManager struct{}

func (fakeTxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fakeConference struct {
	link string
	err  error
}

func (f fakeConference) CreateConferenceLink(_ context.Context, _, _, _ string) (*string, error) {
	if f.err != nil {
		return nil, f.err
	}
	link := f.link
	return &link, nil
}
