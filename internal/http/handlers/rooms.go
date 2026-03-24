package handlers

import (
	"net/http"

	"room-booking-service/internal/http/middleware"
)

// ListRooms godoc
// @Summary Список переговорок
// @Tags Rooms
// @Security BearerAuth
// @Produce json
// @Success 200 {object} handlers.roomsResponse
// @Router /rooms/list [get]
func (h *Handler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.rooms.List(r.Context())
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, roomsResponse{Rooms: rooms})
}

// CreateRoom godoc
// @Summary Создать переговорку
// @Tags Rooms
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body handlers.createRoomRequest true "room"
// @Success 201 {object} handlers.roomResponse
// @Failure 403 {object} map[string]any
// @Router /rooms/create [post]
func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := decodeJSON(r, &req); err != nil {
		middleware.WriteError(w, err)
		return
	}
	room, err := h.rooms.Create(r.Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, roomResponse{Room: *room})
}
