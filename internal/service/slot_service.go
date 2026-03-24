package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"room-booking-service/internal/domain"
	apperrors "room-booking-service/internal/errors"
	"room-booking-service/pkg/timeutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type scheduleRepo interface {
	Create(ctx context.Context, schedule domain.Schedule) (*domain.Schedule, error)
	GetByRoomID(ctx context.Context, roomID string) (*domain.Schedule, error)
	List(ctx context.Context) ([]domain.Schedule, error)
}

type slotRepo interface {
	BulkUpsert(ctx context.Context, slots []domain.Slot) error
	ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]domain.Slot, error)
	GetByID(ctx context.Context, slotID string) (*domain.Slot, error)
}

type SlotGenerator interface {
	GenerateForSchedule(ctx context.Context, schedule domain.Schedule, fromDate time.Time, horizonDays int) error
	ExtendAll(ctx context.Context) error
}

type ScheduleService struct {
	rooms       roomRepo
	schedules   scheduleRepo
	slotGen     SlotGenerator
	horizonDays int
}

func NewScheduleService(rooms roomRepo, schedules scheduleRepo, slotGen SlotGenerator, horizonDays int) *ScheduleService {
	return &ScheduleService{
		rooms:       rooms,
		schedules:   schedules,
		slotGen:     slotGen,
		horizonDays: horizonDays,
	}
}

func (s *ScheduleService) Create(ctx context.Context, roomID string, schedule domain.Schedule) (*domain.Schedule, error) {
	if err := validateUUID(roomID, "roomId"); err != nil {
		return nil, err
	}

	validated, err := normalizeAndValidateSchedule(schedule)
	if err != nil {
		return nil, err
	}

	exists, err := s.rooms.Exists(ctx, roomID)
	if err != nil {
		if isInvalidUUIDError(err) {
			return nil, apperrors.ErrInvalidRequest
		}
		return nil, err
	}
	if !exists {
		return nil, apperrors.ErrRoomNotFound
	}

	validated.ID = uuid.NewString()
	validated.RoomID = roomID

	created, err := s.schedules.Create(ctx, validated)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, apperrors.ErrScheduleExists
		}
		if isInvalidUUIDError(err) {
			return nil, apperrors.ErrInvalidRequest
		}
		return nil, err
	}

	if err := s.slotGen.GenerateForSchedule(ctx, *created, time.Now().UTC(), s.horizonDays); err != nil {
		return nil, err
	}

	return created, nil
}

type SlotService struct {
	rooms       roomRepo
	schedules   scheduleRepo
	slots       slotRepo
	horizonDays int
}

func NewSlotService(rooms roomRepo, schedules scheduleRepo, slots slotRepo, horizonDays int) *SlotService {
	return &SlotService{
		rooms:       rooms,
		schedules:   schedules,
		slots:       slots,
		horizonDays: horizonDays,
	}
}

func (s *SlotService) ExtendAll(ctx context.Context) error {
	schedules, err := s.schedules.List(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, schedule := range schedules {
		if err := s.GenerateForSchedule(ctx, schedule, now, s.horizonDays); err != nil {
			return err
		}
	}

	return nil
}

func (s *SlotService) GenerateForSchedule(ctx context.Context, schedule domain.Schedule, fromDate time.Time, horizonDays int) error {
	if horizonDays <= 0 {
		return nil
	}

	startDate := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, time.UTC)

	hourStart, minuteStart, err := timeutil.ParseHHMM(schedule.StartTime)
	if err != nil {
		return fmt.Errorf("parse schedule start time: %w", err)
	}

	hourEnd, minuteEnd, err := timeutil.ParseHHMM(schedule.EndTime)
	if err != nil {
		return fmt.Errorf("parse schedule end time: %w", err)
	}

	days := make(map[int]struct{}, len(schedule.DaysOfWeek))
	for _, d := range schedule.DaysOfWeek {
		days[d] = struct{}{}
	}

	var slots []domain.Slot
	for dayOffset := 0; dayOffset < horizonDays; dayOffset++ {
		date := startDate.AddDate(0, 0, dayOffset)
		if _, ok := days[timeutil.WeekdayToISO(date.Weekday())]; !ok {
			continue
		}

		slotStart := time.Date(date.Year(), date.Month(), date.Day(), hourStart, minuteStart, 0, 0, time.UTC)
		windowEnd := time.Date(date.Year(), date.Month(), date.Day(), hourEnd, minuteEnd, 0, 0, time.UTC)

		for slotStart.Add(30*time.Minute).Equal(windowEnd) || slotStart.Add(30*time.Minute).Before(windowEnd) {
			slotEnd := slotStart.Add(30 * time.Minute)

			slots = append(slots, domain.Slot{
				ID:     deterministicSlotID(schedule.RoomID, slotStart, slotEnd),
				RoomID: schedule.RoomID,
				Start:  slotStart,
				End:    slotEnd,
			})

			slotStart = slotEnd
		}
	}

	return s.slots.BulkUpsert(ctx, slots)
}

func deterministicSlotID(roomID string, start, end time.Time) string {
	payload := []byte(fmt.Sprintf(
		"%s|%s|%s",
		roomID,
		start.UTC().Format(time.RFC3339),
		end.UTC().Format(time.RFC3339),
	))
	return uuid.NewSHA1(uuid.NameSpaceOID, payload).String()
}

func (s *SlotService) ListAvailableByRoomAndDate(ctx context.Context, roomID string, date string) ([]domain.Slot, error) {
	if err := validateUUID(roomID, "roomId"); err != nil {
		return nil, err
	}

	if date == "" {
		return nil, apperrors.New(
			apperrors.ErrInvalidRequest.Code,
			"date is required",
			http.StatusBadRequest,
		)
	}

	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, apperrors.New(
			apperrors.ErrInvalidRequest.Code,
			"date must be in YYYY-MM-DD format",
			http.StatusBadRequest,
		)
	}

	exists, err := s.rooms.Exists(ctx, roomID)
	if err != nil {
		if isInvalidUUIDError(err) {
			return nil, apperrors.ErrInvalidRequest
		}
		return nil, err
	}
	if !exists {
		return nil, apperrors.ErrRoomNotFound
	}

	_, err = s.schedules.GetByRoomID(ctx, roomID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		if isInvalidUUIDError(err) {
			return nil, apperrors.ErrInvalidRequest
		}
		return nil, err
	}

	dateStart, _ := timeutil.StartEndOfUTCDate(parsed.UTC())

	slots, err := s.slots.ListAvailableByRoomAndDate(ctx, roomID, dateStart)
	if err != nil {
		if isInvalidUUIDError(err) {
			return nil, apperrors.ErrInvalidRequest
		}
		return nil, err
	}

	return slots, nil
}

func normalizeAndValidateSchedule(schedule domain.Schedule) (domain.Schedule, error) {
	if len(schedule.DaysOfWeek) == 0 {
		return schedule, apperrors.New(
			apperrors.ErrInvalidRequest.Code,
			"daysOfWeek must not be empty",
			http.StatusBadRequest,
		)
	}

	set := map[int]struct{}{}
	for _, day := range schedule.DaysOfWeek {
		if day < 1 || day > 7 {
			return schedule, apperrors.New(
				apperrors.ErrInvalidRequest.Code,
				"daysOfWeek must contain values from 1 to 7",
				http.StatusBadRequest,
			)
		}
		set[day] = struct{}{}
	}

	normalizedDays := make([]int, 0, len(set))
	for day := range set {
		normalizedDays = append(normalizedDays, day)
	}
	sort.Ints(normalizedDays)
	schedule.DaysOfWeek = normalizedDays

	start, err := time.Parse("15:04", schedule.StartTime)
	if err != nil {
		return schedule, apperrors.New(
			apperrors.ErrInvalidRequest.Code,
			"invalid startTime",
			http.StatusBadRequest,
		)
	}

	end, err := time.Parse("15:04", schedule.EndTime)
	if err != nil {
		return schedule, apperrors.New(
			apperrors.ErrInvalidRequest.Code,
			"invalid endTime",
			http.StatusBadRequest,
		)
	}

	if !start.Before(end) {
		return schedule, apperrors.New(
			apperrors.ErrInvalidRequest.Code,
			"startTime must be before endTime",
			http.StatusBadRequest,
		)
	}

	if end.Sub(start) < 30*time.Minute {
		return schedule, apperrors.New(
			apperrors.ErrInvalidRequest.Code,
			"schedule window must be at least 30 minutes",
			http.StatusBadRequest,
		)
	}

	return schedule, nil
}