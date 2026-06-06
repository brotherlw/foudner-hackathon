package gateway

import "net/http"

type PublicKeyHandler struct {
	Grants *GrantStore
}

func (h *PublicKeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(h.Grants.PublicKeyBase64()))
}
