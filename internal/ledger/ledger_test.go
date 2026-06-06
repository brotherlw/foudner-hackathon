package ledger

import (
	"os"
	"strings"
	"testing"
)

func TestFileLedgerAppend_writesJSONLWithoutRawIP(t *testing.T) {
	path := t.TempDir() + "/events.jsonl"
	ledger := NewFileLedger(path)

	rawIP := "203.0.113.10"
	if err := ledger.Append(Event{
		Type:         "access_denied",
		ResourcePath: "/api/premium-report",
		Decision:     "denied",
		IPHash:       HashIP(rawIP, "test-salt"),
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if strings.Contains(content, rawIP) {
		t.Fatalf("ledger contains raw IP %q: %s", rawIP, content)
	}
	if !strings.Contains(content, `"ip_hash"`) {
		t.Fatalf("ledger does not contain ip_hash: %s", content)
	}
	if !strings.HasSuffix(content, "\n") {
		t.Fatalf("ledger entry is not newline-delimited: %q", content)
	}
}

func TestMemoryLedgerAppend_recordsEvents(t *testing.T) {
	ledger := NewMemoryLedger()
	events := []Event{
		{Type: "challenge_issued", ResourcePath: "/a", Decision: "denied"},
		{Type: "access_granted", ResourcePath: "/a", Decision: "granted"},
	}
	for _, event := range events {
		if err := ledger.Append(event); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	got := ledger.Events()
	if len(got) != len(events) {
		t.Fatalf("events len = %d, want %d", len(got), len(events))
	}
	for i := range events {
		if got[i].Type != events[i].Type {
			t.Fatalf("event %d type = %q, want %q", i, got[i].Type, events[i].Type)
		}
		if got[i].TS.IsZero() {
			t.Fatalf("event %d timestamp was not set", i)
		}
	}
}
