package handlers

import (
	"net/http"

	"firewall-manager/internal/service"
)

type DashboardHandler struct {
	dashboardSvc *service.DashboardService
}

func NewDashboardHandler(dashboardSvc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardSvc: dashboardSvc}
}

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.dashboardSvc.GetSummary(r.Context())
	if err != nil {
		http.Error(w, "failed to load dashboard summary", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"summary": summary})
}
