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

type EnrollmentHandler struct {
	enrollSvc *service.EnrollmentService
}

func NewEnrollmentHandler(enrollSvc *service.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{enrollSvc: enrollSvc}
}

type createEnrollmentLinkRequest struct {
	ExpiresHours int `json:"expires_hours"`
	MaxUses      int `json:"max_uses"`
}

type acceptEnrollmentRequest struct {
	Token         string `json:"token"`
	Hostname      string `json:"hostname"`
	EmployeeName  string `json:"employee_name"`
	EmployeeEmail string `json:"employee_email"`
	OSType        string `json:"os_type"`
	CurrentIP     string `json:"current_ip"`
	Fingerprint   string `json:"fingerprint"`
	Permission    bool   `json:"permission"`
}

type approveEnrollmentRequest struct {
	DepartmentID *int64 `json:"department_id"`
}

func (h *EnrollmentHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req createEnrollmentLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	item, linkURL, err := h.enrollSvc.GenerateEnrollmentLink(r.Context(), actorUserID, req.ExpiresHours, req.MaxUses)
	if err != nil {
		handleEnrollmentError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"item": item, "link": linkURL})
}

func (h *EnrollmentHandler) AcceptPublic(w http.ResponseWriter, r *http.Request) {
	var req acceptEnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Token == "" {
		req.Token = r.URL.Query().Get("token")
	}
	item, err := h.enrollSvc.AcceptEnrollment(r.Context(), req.Token, models.DeviceEnrollment{
		Hostname:      req.Hostname,
		EmployeeName:  req.EmployeeName,
		EmployeeEmail: req.EmployeeEmail,
		OSType:        req.OSType,
		CurrentIP:     req.CurrentIP,
		Fingerprint:   req.Fingerprint,
		Permission:    req.Permission,
	})
	if err != nil {
		handleEnrollmentError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"item": item, "message": "Enrollment request submitted. Waiting for admin/MD approval."})
}

func (h *EnrollmentHandler) ListEnrollments(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	status := r.URL.Query().Get("status")
	items, err := h.enrollSvc.ListEnrollments(r.Context(), actorUserID, status)
	if err != nil {
		handleEnrollmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *EnrollmentHandler) Approve(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	enrollmentID, err := strconv.ParseInt(chi.URLParam(r, "enrollmentID"), 10, 64)
	if err != nil || enrollmentID <= 0 {
		http.Error(w, "invalid enrollment id", http.StatusBadRequest)
		return
	}
	var req approveEnrollmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	item, err := h.enrollSvc.ApproveEnrollment(r.Context(), actorUserID, enrollmentID, req.DepartmentID)
	if err != nil {
		handleEnrollmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": item})
}

func (h *EnrollmentHandler) Disable(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	enrollmentID, err := strconv.ParseInt(chi.URLParam(r, "enrollmentID"), 10, 64)
	if err != nil || enrollmentID <= 0 {
		http.Error(w, "invalid enrollment id", http.StatusBadRequest)
		return
	}
	if err := h.enrollSvc.DisableEnrollment(r.Context(), actorUserID, enrollmentID); err != nil {
		handleEnrollmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"disabled": true})
}

func (h *EnrollmentHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	items, err := h.enrollSvc.ListNotifications(r.Context(), actorUserID)
	if err != nil {
		handleEnrollmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func handleEnrollmentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrPermissionRequired):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, service.ErrConsentRequired):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrInvalidLink):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, repository.ErrEnrollmentNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
