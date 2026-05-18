package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"firewall-manager/internal/http/middleware"
	"firewall-manager/internal/models"
	"firewall-manager/internal/repository"
	"firewall-manager/internal/service"

	"github.com/go-chi/chi/v5"
)

type PolicyHandler struct {
	policySvc *service.PolicyService
}

func NewPolicyHandler(policySvc *service.PolicyService) *PolicyHandler {
	return &PolicyHandler{policySvc: policySvc}
}

type policyRequest struct {
	Name         string `json:"name"`
	PolicyType   string `json:"policy_type"`
	Action       string `json:"action"`
	Target       string `json:"target"`
	Department   string `json:"department"`
	ScheduleJSON string `json:"schedule_json"`
	IsEnabled    bool   `json:"is_enabled"`
}

func (h *PolicyHandler) List(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	items, err := h.policySvc.List(r.Context(), actorUserID)
	if err != nil {
		handlePolicyError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *PolicyHandler) Create(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req policyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	created, err := h.policySvc.Create(r.Context(), actorUserID, req.toModel())
	if err != nil {
		handlePolicyError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"item": created})
}

func (h *PolicyHandler) Update(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := parsePolicyID(r)
	if err != nil {
		http.Error(w, "invalid policy id", http.StatusBadRequest)
		return
	}

	var req policyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := h.policySvc.Update(r.Context(), actorUserID, id, req.toModel())
	if err != nil {
		handlePolicyError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": updated})
}

func (h *PolicyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := parsePolicyID(r)
	if err != nil {
		http.Error(w, "invalid policy id", http.StatusBadRequest)
		return
	}

	if err := h.policySvc.Delete(r.Context(), actorUserID, id); err != nil {
		handlePolicyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (req policyRequest) toModel() models.FirewallPolicy {
	schedule := req.ScheduleJSON
	if schedule == "" {
		schedule = "{}"
	}

	return models.FirewallPolicy{
		Name:         req.Name,
		PolicyType:   req.PolicyType,
		Action:       req.Action,
		Target:       req.Target,
		Department:   req.Department,
		ScheduleJSON: schedule,
		IsEnabled:    req.IsEnabled,
	}
}

func parsePolicyID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "policyID"), 10, 64)
}

func handlePolicyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrPolicyForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, service.ErrInvalidPolicy):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, repository.ErrPolicyNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, "policy operation failed", http.StatusInternalServerError)
	}
}
