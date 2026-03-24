package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRoomID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/rooms/room-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("roomId", "room-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	if got := roomID(req); got != "room-1" {
		t.Fatalf("expected room-1, got %s", got)
	}
}

func TestBookingID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/bookings/b1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("bookingId", "b1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	if got := bookingID(req); got != "b1" {
		t.Fatalf("expected b1, got %s", got)
	}
}

func TestPageParams_Defaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/bookings/list", nil)

	page, pageSize, err := pageParams(req)
	if err != nil {
		t.Fatalf("pageParams() error = %v", err)
	}
	if page != 1 || pageSize != 20 {
		t.Fatalf("expected defaults 1/20, got %d/%d", page, pageSize)
	}
}

func TestPageParams_InvalidPageSize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/bookings/list?page=1&pageSize=abc", nil)

	_, _, err := pageParams(req)
	if err == nil {
		t.Fatal("expected error for invalid pageSize")
	}
}

func TestNew(t *testing.T) {
	h := New(nil, nil, nil, nil, nil)
	if h == nil {
		t.Fatal("expected handler")
	}
}