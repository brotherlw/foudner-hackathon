package gateway

import (
	"encoding/json"
	"net/http"

	"github.com/agentic-paywall/agentic-paywall/internal/residency"
)

type DataResidencyHandler struct {
	Statement residency.Statement
}

func (h *DataResidencyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.Statement)
}
