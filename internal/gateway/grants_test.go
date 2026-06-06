package gateway

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGrantStore_Ed25519PublicKeyVerification(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	grants := NewGrantStoreWithKey(privateKey, time.Hour, 2)

	token, err := grants.IssueGrant("/api/premium-report", "pay_123", "0.50", "EUR")
	if err != nil {
		t.Fatalf("IssueGrant: %v", err)
	}
	claims, err := VerifyGrantWithPublicKey(grants.PublicKey(), token, "/api/premium-report")
	if err != nil {
		t.Fatalf("VerifyGrantWithPublicKey: %v", err)
	}
	if claims.PaymentID != "pay_123" {
		t.Fatalf("payment_id = %q, want pay_123", claims.PaymentID)
	}
	if claims.Quota != 2 {
		t.Fatalf("quota = %d, want 2", claims.Quota)
	}
}

func TestGrantStore_rejectsTamperedAndPathMismatch(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	grants := NewGrantStoreWithKey(privateKey, time.Hour, 1)
	token, err := grants.IssueGrant("/api/premium-report", "pay_123", "0.50", "EUR")
	if err != nil {
		t.Fatalf("IssueGrant: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		path    string
		wantErr error
	}{
		{
			name:    "path mismatch",
			token:   token,
			path:    "/api/other",
			wantErr: ErrPathMismatch,
		},
		{
			name:    "tampered token",
			token:   token[:len(token)-1] + "A",
			path:    "/api/premium-report",
			wantErr: ErrInvalidGrant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := VerifyGrantWithPublicKey(grants.PublicKey(), tt.token, tt.path)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGrantStore_quotaExhausted(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	grants := NewGrantStoreWithKey(privateKey, time.Hour, 1)
	token, err := grants.IssueGrant("/api/premium-report", "pay_123", "0.50", "EUR")
	if err != nil {
		t.Fatalf("IssueGrant: %v", err)
	}

	if _, err := grants.ConsumeGrant(token, "/api/premium-report"); err != nil {
		t.Fatalf("first ConsumeGrant: %v", err)
	}
	if _, err := grants.ConsumeGrant(token, "/api/premium-report"); !errors.Is(err, ErrQuotaExhausted) {
		t.Fatalf("second ConsumeGrant err = %v, want %v", err, ErrQuotaExhausted)
	}
}

func TestGrantStore_expired(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	grants := NewGrantStoreWithKey(privateKey, -time.Hour, 1)
	token, err := grants.IssueGrant("/api/premium-report", "pay_123", "0.50", "EUR")
	if err != nil {
		t.Fatalf("IssueGrant: %v", err)
	}

	if _, err := grants.VerifyGrant(token, "/api/premium-report"); !errors.Is(err, ErrGrantExpired) {
		t.Fatalf("VerifyGrant err = %v, want %v", err, ErrGrantExpired)
	}
}

func TestGrantStore_concurrentNoOverserve(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	grants := NewGrantStoreWithKey(privateKey, time.Hour, 3)
	token, err := grants.IssueGrant("/api/premium-report", "pay_123", "0.50", "EUR")
	if err != nil {
		t.Fatalf("IssueGrant: %v", err)
	}

	var allowed int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := grants.ConsumeGrant(token, "/api/premium-report"); err == nil {
				atomic.AddInt32(&allowed, 1)
			}
		}()
	}
	wg.Wait()
	if allowed != 3 {
		t.Fatalf("allowed = %d, want 3", allowed)
	}
}

func TestGrantStore_loadOrGeneratePrivateKeyPersists(t *testing.T) {
	path := t.TempDir() + "/grant.key"
	first, err := NewGrantStoreFromKeyFile(path, time.Hour, 1)
	if err != nil {
		t.Fatalf("NewGrantStoreFromKeyFile first: %v", err)
	}
	second, err := NewGrantStoreFromKeyFile(path, time.Hour, 1)
	if err != nil {
		t.Fatalf("NewGrantStoreFromKeyFile second: %v", err)
	}
	if first.PublicKeyBase64() != second.PublicKeyBase64() {
		t.Fatal("public key changed after loading persisted private key")
	}
}

func TestPublicKeyHandler_returnsBase64Key(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	grants := NewGrantStoreWithKey(privateKey, time.Hour, 1)
	handler := &PublicKeyHandler{Grants: grants}

	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent-paywall-key", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	decoded, err := base64.StdEncoding.DecodeString(rec.Body.String())
	if err != nil {
		t.Fatalf("DecodeString: %v", err)
	}
	if len(decoded) != ed25519.PublicKeySize {
		t.Fatalf("public key len = %d, want %d", len(decoded), ed25519.PublicKeySize)
	}
}
