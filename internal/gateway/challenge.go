package gateway

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

const AgentPaywallVersion = 1

type Challenge struct {
	AgentPaywallVersion int              `json:"agent_paywall_version"`
	Error               string           `json:"error"`
	Resource            ResourceInfo     `json:"resource"`
	Accepts             []AcceptOption   `json:"accepts"`
	RetryWith           RetryWith        `json:"retry_with"`
	ExpiresAt           string           `json:"expires_at"`
}

type ResourceInfo struct {
	Path        string `json:"path"`
	Method      string `json:"method"`
	Description string `json:"description"`
}

type AcceptOption struct {
	Scheme              string `json:"scheme"`
	Amount              string `json:"amount"`
	Currency            string `json:"currency"`
	PaymentEndpoint     string `json:"payment_endpoint"`
	GrantVerifyEndpoint string `json:"grant_verify_endpoint"`
}

type RetryWith struct {
	Header      string `json:"header"`
	Description string `json:"description"`
}

func BuildChallenge(baseURL, path, method, description, amount, currency string, ttl time.Duration) Challenge {
	expires := time.Now().UTC().Add(ttl)
	return Challenge{
		AgentPaywallVersion: AgentPaywallVersion,
		Error:               "payment_required",
		Resource: ResourceInfo{
			Path:        path,
			Method:      method,
			Description: description,
		},
		Accepts: []AcceptOption{
			{
				Scheme:              "gateway_fiat",
				Amount:              amount,
				Currency:            currency,
				PaymentEndpoint:     baseURL + "/pay/initiate",
				GrantVerifyEndpoint: baseURL + "/grants/verify",
			},
		},
		RetryWith: RetryWith{
			Header:      "PAYMENT-GRANT",
			Description: "Retry the original request with a valid access grant",
		},
		ExpiresAt: expires.Format(time.RFC3339),
	}
}

func (c Challenge) JSON() ([]byte, error) {
	return json.Marshal(c)
}

func (c Challenge) Base64Header() (string, error) {
	data, err := c.JSON()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func ParseChallenge(data []byte) (Challenge, error) {
	var c Challenge
	err := json.Unmarshal(data, &c)
	return c, err
}
