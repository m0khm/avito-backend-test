package handlers

import (
	"net/http"

	"room-booking-service/internal/domain"
	"room-booking-service/internal/http/middleware"
)

// ListSlots godoc
// @Summary Список доступных слотов
// @Tags Slots
// @Security BearerAuth
// @Produce json
// @Param roomId path string true "room id"
// @Param date query string true "date"
// @Success 200 {object} handlers.slotsResponse
// @Router /rooms/{roomId}/slots/list [get]
func (h *Handler) ListSlots(w http.ResponseWriter, r *http.Request) {
	slots, err := h.slots.ListAvailableByRoomAndDate(r.Context(), roomID(r), r.URL.Query().Get("date"))
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	if slots == nil {
		slots = []domain.Slot{}
	}

	middleware.WriteJSON(w, http.StatusOK, slotsResponse{Slots: slots})
}