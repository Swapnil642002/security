package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"firewall-manager/internal/http/middleware"
	"firewall-manager/internal/service"
)

type AuthHandler struct {
	authSvc        *service.AuthService
	bootstrapToken string
}

func NewAuthHandler(authSvc *service.AuthService, bootstrapToken string) *AuthHandler {
	return &AuthHandler{
		authSvc:        authSvc,
		bootstrapToken: strings.TrimSpace(bootstrapToken),
	}
}

type bootstrapRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) BootstrapAdmin(w http.ResponseWriter, r *http.Request) {
	allowed, err := h.authSvc.IsBootstrapAllowed(r.Context())
	if err != nil {
		http.Error(w, "failed to verify bootstrap status", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "bootstrap is disabled because an admin already exists", http.StatusGone)
		return
	}

	if h.bootstrapToken == "" {
		http.Error(w, "bootstrap token is not configured", http.StatusForbidden)
		return
	}
	if strings.TrimSpace(r.Header.Get("X-Bootstrap-Token")) != h.bootstrapToken {
		http.Error(w, "invalid bootstrap token", http.StatusUnauthorized)
		return
	}

	var req bootstrapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.authSvc.BootstrapAdmin(r.Context(), req.FullName, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminExists):
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.Is(err, service.ErrWeakPassword):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "failed to bootstrap admin", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"user": user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case errors.Is(err, service.ErrInactiveUser):
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, "login failed", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.authSvc.GetCurrentUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to fetch current user", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
