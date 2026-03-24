package handlers

import (
	"net/http"

	"room-booking-service/internal/domain"
	appErrors "room-booking-service/internal/errors"
	"room-booking-service/internal/http/middleware"
)

// CreateSchedule godoc
// @Summary Создать расписание переговорки
// @Tags Schedules
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param roomId path string true "room id"
// @Param request body domain.Schedule true "schedule"
// @Success 201 {object} handlers.scheduleResponse
// @Failure 400 {object} map[string]any
// @Failure 409 {object} map[string]any
// @Router /rooms/{roomId}/schedule/create [post]
func (h *Handler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req createScheduleRequest
	if err := decodeJSON(r, &req); err != nil {
		middleware.WriteError(w, err)
		return
	}

	pathRoomID := roomID(r)
	if req.RoomID == "" || req.RoomID != pathRoomID {
		middleware.WriteError(w, appErrors.New("INVALID_REQUEST", "roomId in path and body must match", http.StatusBadRequest))
		return
	}

	schedule, err := h.schedules.Create(r.Context(), pathRoomID, domain.Schedule{
		RoomID:     req.RoomID,
		DaysOfWeek: req.DaysOfWeek,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
	})
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, scheduleResponse{Schedule: *schedule})
}