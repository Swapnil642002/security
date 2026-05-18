package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"firewall-manager/internal/http/middleware"
	"firewall-manager/internal/models"
	"firewall-manager/internal/service"

	"github.com/go-chi/chi/v5"
)

type FleetHandler struct {
	fleetSvc *service.FleetService
}

func NewFleetHandler(fleetSvc *service.FleetService) *FleetHandler {
	return &FleetHandler{fleetSvc: fleetSvc}
}

type createDepartmentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type createLaptopRequest struct {
	Hostname      string `json:"hostname"`
	EmployeeName  string `json:"employee_name"`
	EmployeeEmail string `json:"employee_email"`
	OSType        string `json:"os_type"`
	DepartmentID  *int64 `json:"department_id"`
	IsActive      bool   `json:"is_active"`
}

type createPolicyAssignmentRequest struct {
	PolicyID       int64  `json:"policy_id"`
	AssignmentType string `json:"assignment_type"`
	DepartmentID   *int64 `json:"department_id"`
	LaptopID       *int64 `json:"laptop_id"`
	IsEnabled      bool   `json:"is_enabled"`
}

func (h *FleetHandler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	items, err := h.fleetSvc.ListDepartments(r.Context(), actorUserID)
	if err != nil {
		handleFleetError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *FleetHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req createDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	d, err := h.fleetSvc.CreateDepartment(r.Context(), actorUserID, req.Name, req.Description)
	if err != nil {
		handleFleetError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"item": d})
}

func (h *FleetHandler) ListLaptops(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	items, err := h.fleetSvc.ListLaptops(r.Context(), actorUserID)
	if err != nil {
		handleFleetError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *FleetHandler) CreateLaptop(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req createLaptopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	item, err := h.fleetSvc.CreateLaptop(r.Context(), actorUserID, models.EmployeeLaptop{
		Hostname:      req.Hostname,
		EmployeeName:  req.EmployeeName,
		EmployeeEmail: req.EmployeeEmail,
		OSType:        req.OSType,
		DepartmentID:  req.DepartmentID,
		IsActive:      req.IsActive,
	})
	if err != nil {
		handleFleetError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"item": item})
}

func (h *FleetHandler) CreatePolicyAssignment(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req createPolicyAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	item, err := h.fleetSvc.CreatePolicyAssignment(r.Context(), actorUserID, models.PolicyAssignment{
		PolicyID:       req.PolicyID,
		AssignmentType: req.AssignmentType,
		DepartmentID:   req.DepartmentID,
		LaptopID:       req.LaptopID,
		IsEnabled:      req.IsEnabled,
	})
	if err != nil {
		handleFleetError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"item": item})
}

func (h *FleetHandler) ListPolicyAssignments(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	policyID, err := strconv.ParseInt(chi.URLParam(r, "policyID"), 10, 64)
	if err != nil || policyID <= 0 {
		http.Error(w, "invalid policy id", http.StatusBadRequest)
		return
	}
	items, err := h.fleetSvc.ListPolicyAssignments(r.Context(), actorUserID, policyID)
	if err != nil {
		handleFleetError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func handleFleetError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrFleetForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
