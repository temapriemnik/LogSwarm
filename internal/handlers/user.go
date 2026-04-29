package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"authbackend/internal/middleware"
	"authbackend/internal/usecase"
)

// UserHandler handles user-related requests.
type UserHandler struct {
	userService usecase.UserService
	authService usecase.AuthService
	router     *mux.Router
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(us usecase.UserService, as usecase.AuthService, router *mux.Router) *UserHandler {
	h := &UserHandler{
		userService: us,
		authService: as,
		router:     router,
	}
	h.RegisterRoutes()
	return h
}

// RegisterRoutes registers the user routes.
func (h *UserHandler) RegisterRoutes() {
	h.router.HandleFunc("/users/me", h.Me).Methods("GET")
	h.router.HandleFunc("/users/password", h.UpdatePassword).Methods("PUT")
}

// Me returns the current user profile.
// @Summary Get current user
// @Description Get currently authenticated user profile
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.User
// @Failure 401 {string} Unauthorized
// @Router /api/users/me [get]
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdatePassword updates the user's password.
// @Summary Update password
// @Description Update password for authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Param old_password body string true "Current password"
// @Param new_password body string true "New password"
// @Security BearerAuth
// @Success 204
// @Failure 400 {string} Invalid request
// @Failure 401 {string} Unauthorized
// @Router /api/users/password [put]
func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.userService.UpdatePassword(r.Context(), user.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}