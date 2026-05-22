package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"firewall-manager/internal/http/middleware"
	"firewall-manager/internal/models"
	"firewall-manager/internal/repository"
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

type agentNextCommandRequest struct {
	AgentToken string `json:"agent_token"`
}

type agentCommandResultRequest struct {
	AgentToken string `json:"agent_token"`
	Success    bool   `json:"success"`
	ResultText string `json:"result_text"`
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

func (h *FleetHandler) DeleteLaptop(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	laptopID, err := strconv.ParseInt(chi.URLParam(r, "laptopID"), 10, 64)
	if err != nil || laptopID <= 0 {
		http.Error(w, "invalid laptop id", http.StatusBadRequest)
		return
	}

	item, err := h.fleetSvc.DeleteLaptop(r.Context(), actorUserID, laptopID)
	if err != nil {
		handleFleetError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"item": item, "deleted": true})
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

func (h *FleetHandler) QueueUSBBlock(w http.ResponseWriter, r *http.Request) {
	h.queueUSBCommand(w, r, true)
}

func (h *FleetHandler) QueueUSBUnblock(w http.ResponseWriter, r *http.Request) {
	h.queueUSBCommand(w, r, false)
}

func (h *FleetHandler) queueUSBCommand(w http.ResponseWriter, r *http.Request, block bool) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	laptopID, err := strconv.ParseInt(chi.URLParam(r, "laptopID"), 10, 64)
	if err != nil || laptopID <= 0 {
		http.Error(w, "invalid laptop id", http.StatusBadRequest)
		return
	}

	item, err := h.fleetSvc.QueueUSBCommand(r.Context(), actorUserID, laptopID, block)
	if err != nil {
		handleFleetError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"item": item})
}

func (h *FleetHandler) AgentClaimNextCommand(w http.ResponseWriter, r *http.Request) {
	token := readAgentToken(r)
	if token == "" {
		var req agentNextCommandRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			token = req.AgentToken
		}
	}

	cmd, err := h.fleetSvc.AgentClaimNextCommand(r.Context(), token)
	if err != nil {
		handleFleetError(w, err)
		return
	}
	if cmd == nil {
		writeJSON(w, http.StatusOK, map[string]any{"item": nil})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": cmd})
}

func (h *FleetHandler) AgentReportCommandResult(w http.ResponseWriter, r *http.Request) {
	commandID, err := strconv.ParseInt(chi.URLParam(r, "commandID"), 10, 64)
	if err != nil || commandID <= 0 {
		http.Error(w, "invalid command id", http.StatusBadRequest)
		return
	}

	var req agentCommandResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.AgentToken == "" {
		req.AgentToken = readAgentToken(r)
	}

	if err := h.fleetSvc.AgentReportCommandResult(r.Context(), req.AgentToken, commandID, req.Success, req.ResultText); err != nil {
		handleFleetError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"updated": true})
}

func readAgentToken(r *http.Request) string {
	token := r.Header.Get("X-Agent-Token")
	if token == "" {
		token = r.URL.Query().Get("agent_token")
	}
	return token
}

func handleFleetError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrFleetForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, service.ErrAgentUnauthorized):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, repository.ErrLaptopNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
