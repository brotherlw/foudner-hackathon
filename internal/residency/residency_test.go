package residency

import (
	"os"
	"strings"
	"testing"

	"github.com/agentic-paywall/agentic-paywall/internal/config"
)

func TestFromConfig_returnsDPOFields(t *testing.T) {
	cfg := config.Default()
	cfg.DataResidency.Region = "eu"
	cfg.DataResidency.StorageLocation = "local append-only JSONL ledger"
	cfg.DataResidency.AllowedSubProcessors = []string{"mollie.com"}

	statement := FromConfig(cfg)
	if statement.Region != "eu" {
		t.Fatalf("region = %q, want eu", statement.Region)
	}
	if statement.RawIPRetained {
		t.Fatal("raw_ip_retained = true, want false")
	}
	if statement.CrossBorderTransfer {
		t.Fatal("cross_border_transfer = true, want false")
	}
	if len(statement.CollectedFields) == 0 {
		t.Fatal("collected_fields is empty")
	}
}

func TestWriteMarkdown_writesLiveArtifact(t *testing.T) {
	path := t.TempDir() + "/DATA-RESIDENCY-LIVE.md"
	statement := FromConfig(config.Default())
	if err := WriteMarkdown(path, statement); err != nil {
		t.Fatalf("WriteMarkdown: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	for _, want := range []string{"Live Data Residency Statement", "raw_ip_retained", "cross_border_transfer"} {
		if !strings.Contains(content, want) {
			t.Fatalf("content missing %q: %s", want, content)
		}
	}
}
