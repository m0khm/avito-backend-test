package handlers

import (
	"net/http"

	"room-booking-service/internal/http/middleware"
)

// Root godoc
// @Summary Service root
// @Tags Health
// @Success 200 {object} map[string]string
// @Router / [get]
func (h *Handler) Root(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Info godoc
// @Summary Service info
// @Tags Health
// @Success 200 {object} map[string]string
// @Router /_info [get]
func (h *Handler) Info(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
