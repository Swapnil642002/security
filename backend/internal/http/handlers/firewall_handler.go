package handlers

import (
	"net/http"

	"firewall-manager/internal/http/middleware"
	"firewall-manager/internal/service"
)

type FirewallHandler struct {
	syncSvc *service.FirewallSyncService
}

func NewFirewallHandler(syncSvc *service.FirewallSyncService) *FirewallHandler {
	return &FirewallHandler{syncSvc: syncSvc}
}

func (h *FirewallHandler) Sync(w http.ResponseWriter, r *http.Request) {
	actorUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	res, err := h.syncSvc.SyncPolicies(r.Context(), actorUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"result": res})
}
