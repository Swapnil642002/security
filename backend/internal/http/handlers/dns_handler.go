package handlers

import (
	"encoding/json"
	"net/http"

	"firewall-manager/internal/dnsfilter"
)

type BlockCategoryRequest struct {
	Category string `json:"category"`
	Block    bool   `json:"block"`
}

func BlockCategoryHandler(w http.ResponseWriter, r *http.Request) {
	var req BlockCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	client := dnsfilter.NextDNSClient{
		ProfileID: "e88957",
		APIKey:    "dc1333f8abdddfffcf6dd08ca14601764f9a3d56",
	}
	err := client.BlockCategory(req.Category, req.Block)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}
