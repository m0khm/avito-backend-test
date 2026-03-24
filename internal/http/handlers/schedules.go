package handlers

import (
	"net/http"

	"room-booking-service/internal/domain"
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
// @Failure 409 {object} map[string]any
// @Router /rooms/{roomId}/schedule/create [post]
func (h *Handler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req domain.Schedule
	if err := decodeJSON(r, &req); err != nil {
		middleware.WriteError(w, err)
		return
	}
	schedule, err := h.schedules.Create(r.Context(), roomID(r), req)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, scheduleResponse{Schedule: *schedule})
}
