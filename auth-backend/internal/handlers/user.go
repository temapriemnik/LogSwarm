package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"authbackend/internal/middleware"
	"authbackend/internal/usecase"
)

type UserHandler struct {
	userService usecase.UserService
	authService usecase.AuthService
	router      *mux.Router
}

func NewUserHandler(us usecase.UserService, as usecase.AuthService, router *mux.Router) *UserHandler {
	h := &UserHandler{
		userService: us,
		authService: as,
		router:     router,
	}
	h.RegisterRoutes()
	return h
}

func (h *UserHandler) RegisterRoutes() {
	h.router.HandleFunc("/users/me", h.Me).Methods("GET")
	h.router.HandleFunc("/users/password", h.UpdatePassword).Methods("PUT")
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

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