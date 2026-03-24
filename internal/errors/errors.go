package errors

import "errors"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

var (
	ErrInvalidRequest    = New("INVALID_REQUEST", "invalid request", 400)
	ErrUnauthorized      = New("UNAUTHORIZED", "unauthorized", 401)
	ErrForbidden         = New("FORBIDDEN", "forbidden", 403)
	ErrNotFound          = New("NOT_FOUND", "not found", 404)
	ErrRoomNotFound      = New("ROOM_NOT_FOUND", "room not found", 404)
	ErrSlotNotFound      = New("SLOT_NOT_FOUND", "slot not found", 404)
	ErrSlotAlreadyBooked = New("SLOT_ALREADY_BOOKED", "slot is already booked", 409)
	ErrBookingNotFound   = New("BOOKING_NOT_FOUND", "booking not found", 404)
	ErrScheduleExists    = New("SCHEDULE_EXISTS", "schedule for this room already exists and cannot be changed", 409)
	ErrInternal          = New("INTERNAL_ERROR", "internal server error", 500)
)

func Is(err error, target *AppError) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == target.Code
	}
	return false
}
