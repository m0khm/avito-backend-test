package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	apperrors "room-booking-service/internal/errors"
)

type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, err error) {
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		appErr = apperrors.ErrInternal
	}
	var payload errorEnvelope
	payload.Error.Code = appErr.Code
	payload.Error.Message = appErr.Message
	WriteJSON(w, appErr.HTTPStatus, payload)
}
