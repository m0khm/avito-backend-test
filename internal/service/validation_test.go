package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"room-booking-service/internal/domain"
	apperrors "room-booking-service/internal/errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestValidateUUID(t *testing.T) {
	if err := validateUUID("", "roomId"); err == nil {
		t.Fatal("expected error for empty uuid")
	}
	if err := validateUUID("not-a-uuid", "roomId"); err == nil {
		t.Fatal("expected error for invalid uuid")
	}
	if err := validateUUID("11111111-1111-1111-1111-111111111111", "roomId"); err != nil {
		t.Fatalf("expected valid uuid, got %v", err)
	}
}

func TestUniqueAndUUIDHelpers(t *testing.T) {
	if !isUniqueViolation(&pgconn.PgError{Code: "23505"}) {
		t.Fatal("expected postgres unique violation")
	}
	if !isUniqueViolation(errors.New("duplicate key value violates unique constraint uq_active_booking_slot")) {
		t.Fatal("expected duplicate key message to be recognized")
	}
	if !isInvalidUUIDError(&pgconn.PgError{Code: "22P02"}) {
		t.Fatal("expected invalid uuid error")
	}
}

func TestSlotService_ListAvailableByRoomAndDate_InvalidRoomID(t *testing.T) {
	t.Parallel()

	svc := NewSlotService(newFakeRoomRepo(), newFakeScheduleRepo(), newFakeSlotRepo(), 7)

	_, err := svc.ListAvailableByRoomAndDate(context.Background(), "not-a-uuid", time.Now().UTC().Format("2006-01-02"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrInvalidRequest.Code {
		t.Fatalf("expected invalid request error, got %s", appErr.Code)
	}
}

func TestBookingService_Create_InvalidIDs(t *testing.T) {
	t.Parallel()

	svc := NewBookingService(newFakeBookingRepo(), newFakeSlotRepo(), fakeTxManager{}, fakeConference{})

	_, err := svc.Create(context.Background(), "not-a-uuid", "also-not-a-uuid", false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrInvalidRequest.Code {
		t.Fatalf("expected invalid request error, got %s", appErr.Code)
	}
}

func TestNormalizeAndValidateSchedule(t *testing.T) {
	tests := []struct {
		name    string
		sched   domain.Schedule
		wantErr bool
		wantLen int
	}{
		{
			name:    "empty days",
			sched:   domain.Schedule{StartTime: "09:00", EndTime: "10:00"},
			wantErr: true,
		},
		{
			name:    "invalid weekday",
			sched:   domain.Schedule{DaysOfWeek: []int{0}, StartTime: "09:00", EndTime: "10:00"},
			wantErr: true,
		},
		{
			name:    "invalid start time",
			sched:   domain.Schedule{DaysOfWeek: []int{1}, StartTime: "bad", EndTime: "10:00"},
			wantErr: true,
		},
		{
			name:    "end before start",
			sched:   domain.Schedule{DaysOfWeek: []int{1}, StartTime: "10:00", EndTime: "09:00"},
			wantErr: true,
		},
		{
			name:    "window shorter than 30m",
			sched:   domain.Schedule{DaysOfWeek: []int{1}, StartTime: "10:00", EndTime: "10:15"},
			wantErr: true,
		},
		{
			name:    "deduplicates and sorts days",
			sched:   domain.Schedule{DaysOfWeek: []int{5, 1, 5, 3}, StartTime: "09:00", EndTime: "10:00"},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAndValidateSchedule(tt.sched)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.DaysOfWeek) != tt.wantLen {
				t.Fatalf("expected %d unique days, got %v", tt.wantLen, got.DaysOfWeek)
			}
			if got.DaysOfWeek[0] != 1 || got.DaysOfWeek[1] != 3 || got.DaysOfWeek[2] != 5 {
				t.Fatalf("expected sorted days [1 3 5], got %v", got.DaysOfWeek)
			}
		})
	}
}
