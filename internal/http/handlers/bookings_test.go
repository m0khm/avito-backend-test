package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"room-booking-service/internal/domain"
	"room-booking-service/internal/http/middleware"
	"room-booking-service/internal/service"
	"room-booking-service/pkg/jwtutil"
)

type bookingsTestRepo struct {
	listAllBookings  []domain.Booking
	listAllTotal     int
	futureBookings   []domain.Booking
	getByIDBooking   *domain.Booking
	cancelledBooking *domain.Booking
}

func (r *bookingsTestRepo) Create(ctx context.Context, booking domain.Booking) (*domain.Booking, error) {
	copy := booking
	return &copy, nil
}

func (r *bookingsTestRepo) UpdateConferenceLink(ctx context.Context, bookingID string, conferenceLink *string) error {
	if r.getByIDBooking != nil {
		r.getByIDBooking.ConferenceLink = conferenceLink
	}
	return nil
}

func (r *bookingsTestRepo) ListAll(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	return r.listAllBookings, r.listAllTotal, nil
}

func (r *bookingsTestRepo) ListFutureByUser(ctx context.Context, userID string) ([]domain.Booking, error) {
	return r.futureBookings, nil
}

func (r *bookingsTestRepo) GetByID(ctx context.Context, bookingID string) (*domain.Booking, error) {
	return r.getByIDBooking, nil
}

func (r *bookingsTestRepo) Cancel(ctx context.Context, bookingID string) (*domain.Booking, error) {
	if r.cancelledBooking != nil {
		return r.cancelledBooking, nil
	}
	copy := *r.getByIDBooking
	copy.Status = domain.BookingStatusCancelled
	return &copy, nil
}

func withUserAuth(t *testing.T, h http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	token, err := jwtutil.IssueToken("secret", "user-1", "user", time.Hour)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()

	mw := middleware.NewAuthMiddleware("secret")
	mw.RequireAuth(h).ServeHTTP(rec, req)

	return rec
}

func TestCreateBooking_InvalidBody(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewReader([]byte(`{bad json}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateBooking(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestListBookings_InvalidPage(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodGet, "/bookings/list?page=abc&pageSize=20", nil)
	rec := httptest.NewRecorder()

	h.ListBookings(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestListBookings_Success(t *testing.T) {
	repo := &bookingsTestRepo{
		listAllBookings: []domain.Booking{
			{ID: "b1", SlotID: "s1", UserID: "u1", Status: domain.BookingStatusActive},
		},
		listAllTotal: 1,
	}
	h := &Handler{
		bookings: service.NewBookingService(repo, nil, nil, nil),
	}

	req := httptest.NewRequest(http.MethodGet, "/bookings/list?page=1&pageSize=20", nil)
	rec := httptest.NewRecorder()

	h.ListBookings(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"bookings"`) {
		t.Fatalf("expected bookings in response, body=%s", rec.Body.String())
	}
}

func TestListMyBookings_Success(t *testing.T) {
	repo := &bookingsTestRepo{
		futureBookings: []domain.Booking{
			{ID: "b1", SlotID: "s1", UserID: "user-1", Status: domain.BookingStatusActive},
		},
	}
	h := &Handler{
		bookings: service.NewBookingService(repo, nil, nil, nil),
	}

	rec := withUserAuth(t, http.HandlerFunc(h.ListMyBookings), http.MethodGet, "/bookings/my", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"user-1"`) {
		t.Fatalf("expected user booking in response, body=%s", rec.Body.String())
	}
}

func TestCancelBooking_Success(t *testing.T) {
	repo := &bookingsTestRepo{
		getByIDBooking: &domain.Booking{
			ID:     "booking-1",
			SlotID: "slot-1",
			UserID: "user-1",
			Status: domain.BookingStatusActive,
		},
		cancelledBooking: &domain.Booking{
			ID:     "booking-1",
			SlotID: "slot-1",
			UserID: "user-1",
			Status: domain.BookingStatusCancelled,
		},
	}
	h := &Handler{
		bookings: service.NewBookingService(repo, nil, nil, nil),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("bookingId", "booking-1")
		req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		h.CancelBooking(w, req)
	})

	rec := withUserAuth(t, handler, http.MethodPost, "/bookings/booking-1/cancel", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"cancelled"`) {
		t.Fatalf("expected cancelled status, body=%s", rec.Body.String())
	}
}