package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"authbackend/internal/usecase"
)

type AuthHandler struct {
	authService usecase.AuthService
	router     *mux.Router
}

func NewAuthHandler(as usecase.AuthService, router *mux.Router) *AuthHandler {
	h := &AuthHandler{
		authService: as,
		router:     router,
	}
	h.RegisterRoutes()
	return h
}

func (h *AuthHandler) RegisterRoutes() {
	h.router.HandleFunc("/auth/register", h.Register).Methods("POST")
	h.router.HandleFunc("/auth/login", h.Login).Methods("POST")
	h.router.HandleFunc("/auth/refresh", h.Refresh).Methods("POST")
}

// @Summary Register
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param name body string true "User name"
// @Param email body string true "User email"
// @Param password body string true "User password"
// @Success 201 {object} domain.User
// @Failure 400 {string} Invalid request
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// @Summary Login
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param email body string true "User email"
// @Param password body string true "User password"
// @Success 200 {object} domain.TokenPair
// @Failure 401 {string} Invalid credentials
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokens, user, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user":        user,
	})
}

// @Summary Refresh
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body string true "Refresh token"
// @Success 200 {object} domain.TokenPair
// @Failure 401 {string} Invalid token
// @Router /api/auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokens, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}