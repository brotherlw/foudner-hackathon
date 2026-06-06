package gateway

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidGrant = errors.New("invalid grant")
	ErrGrantExpired = errors.New("grant expired")
	ErrPathMismatch = errors.New("grant path mismatch")
)

type GrantClaims struct {
	ResourcePath string `json:"resource_path"`
	PaymentID    string `json:"payment_id"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	jwt.RegisteredClaims
}

type GrantStore struct {
	secret  []byte
	ttl     time.Duration
	pending map[string]string // payment_id -> grant token
	mu      sync.RWMutex
}

func NewGrantStore(secret string, ttl time.Duration) *GrantStore {
	return &GrantStore{
		secret:  []byte(secret),
		ttl:     ttl,
		pending: make(map[string]string),
	}
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
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
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
	raw, err := decodeGrantToken(tokenB64)
	if err != nil {
		return nil, ErrInvalidGrant
	}

	token, err := jwt.ParseWithClaims(raw, &GrantClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
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
