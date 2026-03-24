package handlers

import (
	"net/http"

	"room-booking-service/internal/http/middleware"
)

// DummyLogin godoc
// @Summary Получить тестовый JWT по роли
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body handlers.dummyLoginRequest true "role"
// @Success 200 {object} handlers.tokenResponse
// @Failure 400 {object} map[string]any
// @Router /dummyLogin [post]
func (h *Handler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dummyLoginRequest
	if err := decodeJSON(r, &req); err != nil {
		middleware.WriteError(w, err)
		return
	}
	token, err := h.auth.DummyLogin(r.Context(), req.Role)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, tokenResponse{Token: token})
}

// Register godoc
// @Summary Регистрация пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body handlers.registerRequest true "register"
// @Success 201 {object} handlers.userResponse
// @Failure 400 {object} map[string]any
// @Router /register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		middleware.WriteError(w, err)
		return
	}
	user, err := h.auth.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusCreated, userResponse{User: *user})
}

// Login godoc
// @Summary Авторизация по email и паролю
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body handlers.loginRequest true "login"
// @Success 200 {object} handlers.tokenResponse
// @Failure 401 {object} map[string]any
// @Router /login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		middleware.WriteError(w, err)
		return
	}
	token, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	middleware.WriteJSON(w, http.StatusOK, tokenResponse{Token: token})
}
