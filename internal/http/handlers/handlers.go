package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"room-booking-service/internal/domain"
	apperrors "room-booking-service/internal/errors"
	"room-booking-service/internal/http/middleware"
	"room-booking-service/internal/service"
)

type Handler struct {
	auth      *service.AuthService
	rooms     *service.RoomService
	schedules *service.ScheduleService
	slots     *service.SlotService
	bookings  *service.BookingService
}

func New(auth *service.AuthService, rooms *service.RoomService, schedules *service.ScheduleService, slots *service.SlotService, bookings *service.BookingService) *Handler {
	return &Handler{auth: auth, rooms: rooms, schedules: schedules, slots: slots, bookings: bookings}
}

type tokenResponse struct {
	Token string `json:"token"`
}

type userResponse struct {
	User domain.User `json:"user"`
}

type roomResponse struct {
	Room domain.Room `json:"room"`
}

type roomsResponse struct {
	Rooms []domain.Room `json:"rooms"`
}

type scheduleResponse struct {
	Schedule domain.Schedule `json:"schedule"`
}

type slotsResponse struct {
	Slots []domain.Slot `json:"slots"`
}

type bookingResponse struct {
	Booking domain.Booking `json:"booking"`
}

type bookingsResponse struct {
	Bookings   []domain.Booking   `json:"bookings"`
	Pagination *domain.Pagination `json:"pagination,omitempty"`
}

type dummyLoginRequest struct {
	Role string `json:"role"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type createRoomRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Capacity    *int    `json:"capacity"`
}

type createScheduleRequest struct {
	RoomID     string `json:"roomId"`
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type createBookingRequest struct {
	SlotID               string `json:"slotId"`
	CreateConferenceLink bool   `json:"createConferenceLink"`
}

func roomID(r *http.Request) string {
	return chi.URLParam(r, "roomId")
}

func bookingID(r *http.Request) string {
	return chi.URLParam(r, "bookingId")
}

func userID(r *http.Request) string {
	return middleware.UserIDFromContext(r.Context())
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return apperrors.New(apperrors.ErrInvalidRequest.Code, "invalid request body", http.StatusBadRequest)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return apperrors.New(apperrors.ErrInvalidRequest.Code, "request body must contain a single JSON object", http.StatusBadRequest)
	}
	return nil
}

func pageParams(r *http.Request) (int, int, error) {
	page := 1
	pageSize := 20
	if raw := r.URL.Query().Get("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return 0, 0, apperrors.New(apperrors.ErrInvalidRequest.Code, "page must be integer", http.StatusBadRequest)
		}
		page = parsed
	}
	if raw := r.URL.Query().Get("pageSize"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return 0, 0, apperrors.New(apperrors.ErrInvalidRequest.Code, "pageSize must be integer", http.StatusBadRequest)
		}
		pageSize = parsed
	}
	return page, pageSize, nil
}
