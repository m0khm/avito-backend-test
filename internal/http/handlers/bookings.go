package handlers

import (
	"net/http"

	"room-booking-service/internal/http/middleware"
)

// CreateBooking godoc
// @Summary Создать бронь
// @Tags Bookings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body handlers.createBookingRequest true "booking"
// @Success 201 {object} handlers.bookingResponse
// @Failure 409 {object} map[string]any
// @Router /bookings/create [post]
func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var req createBookingRequest
	if err := decodeJSON(r, &req); err != nil {
		middleware.WriteError(w, err)
		return
	}
	booking, err := h.bookings.Create(r.Context(), req.SlotID, userID(r), req.CreateConferenceLink)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, bookingResponse{Booking: *booking})
}

// ListBookings godoc
// @Summary Список всех броней
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Success 200 {object} handlers.bookingsResponse
// @Router /bookings/list [get]
func (h *Handler) ListBookings(w http.ResponseWriter, r *http.Request) {
	page, pageSize, err := pageParams(r)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	bookings, pagination, err := h.bookings.ListAll(r.Context(), page, pageSize)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, bookingsResponse{Bookings: bookings, Pagination: &pagination})
}

// ListMyBookings godoc
// @Summary Список броней текущего пользователя
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Success 200 {object} handlers.bookingsResponse
// @Router /bookings/my [get]
func (h *Handler) ListMyBookings(w http.ResponseWriter, r *http.Request) {
	bookings, err := h.bookings.ListFutureByUser(r.Context(), userID(r))
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, bookingsResponse{Bookings: bookings})
}

// CancelBooking godoc
// @Summary Отменить бронь
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Param bookingId path string true "booking id"
// @Success 200 {object} handlers.bookingResponse
// @Router /bookings/{bookingId}/cancel [post]
func (h *Handler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	booking, err := h.bookings.Cancel(r.Context(), bookingID(r), userID(r))
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, bookingResponse{Booking: *booking})
}
