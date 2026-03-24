package service

import (
    "errors"
    "strings"

    apperrors "room-booking-service/internal/errors"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgconn"
)

func validateUUID(value, field string) error {
    if strings.TrimSpace(value) == "" {
        return apperrors.New(apperrors.ErrInvalidRequest.Code, field+" is required", 400)
    }
    if _, err := uuid.Parse(value); err != nil {
        return apperrors.New(apperrors.ErrInvalidRequest.Code, field+" must be a valid UUID", 400)
    }
    return nil
}

func isUniqueViolation(err error) bool {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return pgErr.Code == "23505"
    }
    msg := strings.ToLower(err.Error())
    return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "uq_active_booking_slot")
}

func isInvalidUUIDError(err error) bool {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return pgErr.Code == "22P02"
    }
    return false
}
