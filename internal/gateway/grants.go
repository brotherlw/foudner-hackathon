package gateway

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidGrant   = errors.New("invalid grant")
	ErrGrantExpired   = errors.New("grant expired")
	ErrPathMismatch   = errors.New("grant path mismatch")
	ErrQuotaExhausted = errors.New("grant quota exhausted")
)

type GrantClaims struct {
	ResourcePath string `json:"resource_path"`
	PaymentID    string `json:"payment_id"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	Quota        int    `json:"quota"`
	jwt.RegisteredClaims
}

type GrantStore struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	ttl        time.Duration
	quota      int
	pending    map[string]string // payment_id -> grant token
	quotas     map[string]quotaState
	mu         sync.RWMutex
}

type quotaState struct {
	remaining int
	expiresAt time.Time
}

func NewGrantStore(secret string, ttl time.Duration) *GrantStore {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	return NewGrantStoreWithKey(privateKey, ttl, 1)
}

func NewGrantStoreWithKey(privateKey ed25519.PrivateKey, ttl time.Duration, quota int) *GrantStore {
	if quota <= 0 {
		quota = 1
	}
	publicKey := privateKey.Public().(ed25519.PublicKey)
	return &GrantStore{
		privateKey: privateKey,
		publicKey:  publicKey,
		ttl:        ttl,
		quota:      quota,
		pending:    make(map[string]string),
		quotas:     make(map[string]quotaState),
	}
}

func NewGrantStoreFromKeyFile(path string, ttl time.Duration, quota int) (*GrantStore, error) {
	privateKey, err := loadOrGeneratePrivateKey(path)
	if err != nil {
		return nil, err
	}
	return NewGrantStoreWithKey(privateKey, ttl, quota), nil
}

func (s *GrantStore) PublicKey() ed25519.PublicKey {
	return append(ed25519.PublicKey(nil), s.publicKey...)
}

func (s *GrantStore) PublicKeyBase64() string {
	return base64.StdEncoding.EncodeToString(s.publicKey)
}

func (s *GrantStore) Quota() int {
	return s.quota
}

func (s *GrantStore) IssueGrant(resourcePath, paymentID, amount, currency string) (string, error) {
	jti, err := randomID()
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	claims := GrantClaims{
		ResourcePath: resourcePath,
		PaymentID:    paymentID,
		Amount:       amount,
		Currency:     currency,
		Quota:        s.quota,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString([]byte(signed)), nil
}

func (s *GrantStore) StorePendingGrant(paymentID, grantToken string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[paymentID] = grantToken
}

func (s *GrantStore) GetPendingGrant(paymentID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	grant, ok := s.pending[paymentID]
	return grant, ok
}

func (s *GrantStore) VerifyGrant(tokenB64, requestPath string) (*GrantClaims, error) {
	return VerifyGrantWithPublicKey(s.publicKey, tokenB64, requestPath)
}

func (s *GrantStore) ConsumeGrant(tokenB64, requestPath string) (*GrantClaims, error) {
	claims, err := s.VerifyGrant(tokenB64, requestPath)
	if err != nil {
		return nil, err
	}
	if claims.ID == "" {
		return nil, ErrInvalidGrant
	}
	expiresAt := time.Now().UTC().Add(s.ttl)
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}
	quota := claims.Quota
	if quota <= 0 {
		quota = 1
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneExpiredQuotas(time.Now().UTC())
	state, ok := s.quotas[claims.ID]
	if !ok {
		state = quotaState{remaining: quota, expiresAt: expiresAt}
	}
	if state.remaining <= 0 {
		s.quotas[claims.ID] = state
		return nil, ErrQuotaExhausted
	}
	state.remaining--
	s.quotas[claims.ID] = state
	return claims, nil
}

func (s *GrantStore) pruneExpiredQuotas(now time.Time) {
	for jti, state := range s.quotas {
		if !state.expiresAt.IsZero() && state.expiresAt.Before(now) {
			delete(s.quotas, jti)
		}
	}
}

func VerifyGrantWithPublicKey(pub ed25519.PublicKey, tokenB64, requestPath string) (*GrantClaims, error) {
	raw, err := decodeGrantToken(tokenB64)
	if err != nil {
		return nil, ErrInvalidGrant
	}

	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.ParseWithClaims(raw, &GrantClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodEdDSA {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return pub, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidGrant
	}

	claims, ok := token.Claims.(*GrantClaims)
	if !ok {
		return nil, ErrInvalidGrant
	}
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrGrantExpired
	}
	if claims.ResourcePath != requestPath {
		return nil, ErrPathMismatch
	}
	return claims, nil
}

func decodeGrantToken(tokenB64 string) (string, error) {
	if decoded, err := base64.RawURLEncoding.DecodeString(tokenB64); err == nil {
		return string(decoded), nil
	}
	if decoded, err := base64.URLEncoding.DecodeString(tokenB64); err == nil {
		return string(decoded), nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(tokenB64); err == nil {
		return string(decoded), nil
	}
	return tokenB64, nil
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func loadOrGeneratePrivateKey(path string) (ed25519.PrivateKey, error) {
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
			if err != nil {
				return nil, fmt.Errorf("decode grant private key: %w", err)
			}
			if len(decoded) != ed25519.PrivateKeySize {
				return nil, fmt.Errorf("grant private key has %d bytes, want %d", len(decoded), ed25519.PrivateKeySize)
			}
			return ed25519.PrivateKey(decoded), nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read grant private key: %w", err)
		}
	}

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return privateKey, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(privateKey)
	if err := os.WriteFile(path, []byte(encoded), 0o600); err != nil {
		return nil, err
	}
	return privateKey, nil
}
