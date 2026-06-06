package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Event struct {
	TS           time.Time `json:"ts"`
	Type         string    `json:"type"`
	PaymentID    string    `json:"payment_id,omitempty"`
	ResourcePath string    `json:"resource_path,omitempty"`
	Amount       string    `json:"amount,omitempty"`
	Currency     string    `json:"currency,omitempty"`
	AgentID      string    `json:"agent_id,omitempty"`
	Decision     string    `json:"decision,omitempty"`
	IPHash       string    `json:"ip_hash,omitempty"`
}

type Ledger interface {
	Append(Event) error
}

type NopLedger struct{}

func (NopLedger) Append(Event) error {
	return nil
}

type FileLedger struct {
	path string
	mu   sync.Mutex
}

func NewFileLedger(path string) *FileLedger {
	return &FileLedger{path: path}
}

func (l *FileLedger) Append(event Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if event.TS.IsZero() {
		event.TS = time.Now().UTC()
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

type MemoryLedger struct {
	mu     sync.Mutex
	events []Event
}

func NewMemoryLedger() *MemoryLedger {
	return &MemoryLedger{}
}

func (l *MemoryLedger) Append(event Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if event.TS.IsZero() {
		event.TS = time.Now().UTC()
	}
	l.events = append(l.events, event)
	return nil
}

func (l *MemoryLedger) Events() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	events := make([]Event, len(l.events))
	copy(events, l.events)
	return events
}

func HashIP(ip, salt string) string {
	if ip == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(salt + ":" + ip))
	return hex.EncodeToString(sum[:])
}
