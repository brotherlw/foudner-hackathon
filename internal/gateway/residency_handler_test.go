package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentic-paywall/agentic-paywall/internal/config"
	"github.com/agentic-paywall/agentic-paywall/internal/residency"
)

func TestDataResidencyHandler_returnsStatement(t *testing.T) {
	statement := residency.FromConfig(config.Default())
	handler := &DataResidencyHandler{Statement: statement}

	req := httptest.NewRequest(http.MethodGet, "/.well-known/data-residency", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got residency.Statement
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode statement: %v", err)
	}
	if got.RawIPRetained {
		t.Fatal("raw_ip_retained = true, want false")
	}
	if got.CrossBorderTransfer {
		t.Fatal("cross_border_transfer = true, want false")
	}
	if got.Region != statement.Region {
		t.Fatalf("region = %q, want %q", got.Region, statement.Region)
	}
}
